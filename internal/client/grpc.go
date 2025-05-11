package client

import (
	"context"
	"fmt"

	pb "Deist_Calc/internal/proto"
	"Deist_Calc/internal/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	client pb.CalculatorServiceClient
}

func NewGRPCClient(serverAddr string) (*GRPCClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к серверу: %v", err)
	}

	client := pb.NewCalculatorServiceClient(conn)
	return &GRPCClient{client: client}, nil
}

func (c *GRPCClient) Calculate(expression, userID string) (string, error) {
	resp, err := c.client.Calculate(context.Background(), &pb.CalculateRequest{
		Expression: expression,
		UserId:     userID,
	})
	if err != nil {
		return "", fmt.Errorf("ошибка вызова Calculate: %v", err)
	}
	return resp.Result, nil
}

func (c *GRPCClient) GetExpressions(userID string) ([]storage.Expression, error) {
	resp, err := c.client.GetExpressions(context.Background(), &pb.GetExpressionsRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка вызова GetExpressions: %v", err)
	}

	var expressions []storage.Expression
	for _, expr := range resp.Expressions {
		expressions = append(expressions, storage.Expression{
			ID:         expr.Id,
			Expression: expr.Expression,
			Result:     expr.Result,
			Status:     expr.Status,
			CreatedAt:  expr.CreatedAt,
		})
	}
	return expressions, nil
}
