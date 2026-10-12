package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"orders-clean-architecture/internal/domain"
	"orders-clean-architecture/internal/usecase"
)

type Handler struct {
	list   *usecase.ListOrders
	create *usecase.CreateOrder
}

func NewHandler(list *usecase.ListOrders, create *usecase.CreateOrder) *Handler {
	return &Handler{list: list, create: create}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch {
	case r.URL.Path == "/order" && r.Method == http.MethodGet:
		ctx, cancel := contextWithTimeout(r)
		defer cancel()
		orders, err := h.list.Execute(ctx)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		_ = json.NewEncoder(w).Encode(orders)
	case (r.URL.Path == "/order" || r.URL.Path == "/order/") && r.Method == http.MethodPost:
		ctx, cancel := contextWithTimeout(r)
		defer cancel()
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var input struct {
			Customer    string  `json:"customer"`
			Description string  `json:"description"`
			Amount      float64 `json:"amount"`
		}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&input); err != nil && err != io.EOF {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		order, err := h.create.Execute(ctx, domain.Order{Customer: input.Customer, Description: input.Description, Amount: input.Amount})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(order)
	case r.URL.Path == "/order" && r.Method != http.MethodGet && r.Method != http.MethodPost:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	case strings.HasPrefix(r.URL.Path, "/order/"):
		idText := strings.TrimPrefix(r.URL.Path, "/order/")
		id, err := strconv.ParseInt(idText, 10, 64)
		if err != nil || id <= 0 {
			writeError(w, http.StatusBadRequest, "invalid order id")
			return
		}
		http.NotFound(w, r)
	default:
		http.NotFound(w, r)
	}
}

func contextWithTimeout(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 5*time.Second)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
