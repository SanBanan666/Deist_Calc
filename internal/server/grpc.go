package server

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "Deist_Calc/internal/proto"
	"Deist_Calc/internal/storage"

	"google.golang.org/grpc"
)

type CalculatorServer struct {
	pb.UnimplementedCalculatorServiceServer
	storage *storage.Storage
}

func NewCalculatorServer(storage *storage.Storage) *CalculatorServer {
	return &CalculatorServer{
		storage: storage,
	}
}

func (s *CalculatorServer) Calculate(ctx context.Context, req *pb.CalculateRequest) (*pb.CalculateResponse, error) {
	// Сохраняем выражение в базу данных
	exprID, err := s.storage.SaveExpression(req.UserId, req.Expression)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения выражения: %v", err)
	}

	return &pb.CalculateResponse{
		Result: exprID,
	}, nil
}

func (s *CalculatorServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	// Получаем задачу из базы данных
	expr, err := s.storage.GetPendingExpression()
	if err != nil {
		return nil, fmt.Errorf("нет доступных задач")
	}

	return &pb.Task{
		Id:         expr.ID,
		Expression: expr.Expression,
		UserId:     expr.UserID,
		Status:     expr.Status,
	}, nil
}

func (s *CalculatorServer) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.UpdateTaskResponse, error) {
	// Обновляем статус задачи в базе данных
	err := s.storage.UpdateExpression(req.TaskId, req.Result, "completed")
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления выражения: %v", err)
	}

	return &pb.UpdateTaskResponse{
		Success: true,
	}, nil
}

func (s *CalculatorServer) GetExpressions(ctx context.Context, req *pb.GetExpressionsRequest) (*pb.GetExpressionsResponse, error) {
	expressions, err := s.storage.GetUserExpressions(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения выражений: %v", err)
	}

	var pbExpressions []*pb.Expression
	for _, expr := range expressions {
		pbExpressions = append(pbExpressions, &pb.Expression{
			Id:         expr.ID,
			Expression: expr.Expression,
			Result:     expr.Result,
			Status:     expr.Status,
			CreatedAt:  expr.CreatedAt,
		})
	}

	return &pb.GetExpressionsResponse{
		Expressions: pbExpressions,
	}, nil
}

func StartGRPCServer(storage *storage.Storage) error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("ошибка создания слушателя: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterCalculatorServiceServer(server, NewCalculatorServer(storage))

	log.Println("gRPC сервер запущен на порту 50051")
	return server.Serve(lis)
}
