package rpc

// Este serviço gRPC usa mensagens protobuf compatíveis com o serviço definido
// em proto/orders.proto. O arquivo gerado é incluído no projeto para manter
// o contrato explícito; a implementação delega ao mesmo ListOrdersUseCase.

import (
	"context"

	"google.golang.org/grpc"
	pb "orders-clean-architecture/internal/adapter/rpc/pb"
	"orders-clean-architecture/internal/usecase"
)

type server struct {
	pb.UnimplementedOrdersServiceServer
	list *usecase.ListOrders
}

// Register registers the OrdersService server to the gRPC server.
func Register(grpcServer *grpc.Server, list *usecase.ListOrders) {
	pb.RegisterOrdersServiceServer(grpcServer, &server{list: list})
}

func (s *server) ListOrders(ctx context.Context, _ *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders, err := s.list.Execute(ctx)
	if err != nil {
		return nil, err
	}
	response := &pb.ListOrdersResponse{Orders: make([]*pb.Order, 0, len(orders))}
	for _, order := range orders {
		response.Orders = append(response.Orders, &pb.Order{
			Id:          order.ID,
			Customer:    order.Customer,
			Description: order.Description,
			Amount:      order.Amount,
			CreatedAt:   order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return response, nil
}
