package graphqlapi

import (
	"encoding/json"
	"net/http"

	"github.com/graphql-go/graphql"
	"orders-clean-architecture/internal/domain"
	"orders-clean-architecture/internal/usecase"
)

func NewSchema(list *usecase.ListOrders, create *usecase.CreateOrder) (graphql.Schema, error) {
	orderType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Order",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.Int},
			"customer":    &graphql.Field{Type: graphql.String},
			"description": &graphql.Field{Type: graphql.String},
			"amount":      &graphql.Field{Type: graphql.Float},
			"created_at": &graphql.Field{Type: graphql.String, Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				if order, ok := p.Source.(domain.Order); ok {
					return order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
				}
				return "", nil
			}},
		},
	})
	query := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"ListOrders": &graphql.Field{
				Type:    graphql.NewList(orderType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) { return list.Execute(p.Context) },
			},
		},
	})
	mutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"CreateOrder": &graphql.Field{
				Type: orderType,
				Args: graphql.FieldConfigArgument{
					"customer":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"description": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"amount":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Float)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return create.Execute(p.Context, domain.Order{
						Customer:    p.Args["customer"].(string),
						Description: p.Args["description"].(string),
						Amount:      p.Args["amount"].(float64),
					})
				},
			},
		},
	})
	return graphql.NewSchema(graphql.SchemaConfig{Query: query, Mutation: mutation})
}

func Handler(schema graphql.Schema) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "use POST with a GraphQL JSON body", http.StatusMethodNotAllowed)
			return
		}
		var input struct {
			Query     string                 `json:"query"`
			Variables map[string]interface{} `json:"variables"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		result := graphql.Do(graphql.Params{Schema: schema, RequestString: input.Query, VariableValues: input.Variables, Context: r.Context()})
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(result)
	})
}
