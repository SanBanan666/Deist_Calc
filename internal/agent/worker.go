package agent

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	pb "Deist_Calc/internal/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Worker struct {
	client pb.CalculatorServiceClient
}

func NewWorker(serverAddr string) (*Worker, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к серверу: %v", err)
	}

	client := pb.NewCalculatorServiceClient(conn)
	return &Worker{client: client}, nil
}

func (w *Worker) Start() {
	for {
		task, err := w.client.GetTask(context.Background(), &pb.GetTaskRequest{})
		if err != nil {
			log.Printf("Ошибка получения задачи: %v", err)
			time.Sleep(time.Second)
			continue
		}

		result := compute(task.Expression)
		_, err = w.client.UpdateTask(context.Background(), &pb.UpdateTaskRequest{
			TaskId: task.Id,
			Result: result,
		})
		if err != nil {
			log.Printf("Ошибка обновления задачи: %v", err)
		}
	}
}

func compute(expression string) string {
	tokens := strings.Fields(expression)
	if len(tokens) < 3 {
		return "Ошибка: неверный формат выражения"
	}

	// Сначала обрабатываем умножение и деление
	for i := 1; i < len(tokens)-1; i += 2 {
		if tokens[i] == "*" || tokens[i] == "/" {
			a, err := strconv.ParseFloat(tokens[i-1], 64)
			if err != nil {
				return "Ошибка: неверный формат числа"
			}
			b, err := strconv.ParseFloat(tokens[i+1], 64)
			if err != nil {
				return "Ошибка: неверный формат числа"
			}

			var result float64
			switch tokens[i] {
			case "*":
				result = a * b
			case "/":
				if b == 0 {
					return "Ошибка: деление на ноль"
				}
				result = a / b
			}

			// Заменяем обработанную часть на результат
			if result == float64(int64(result)) {
				tokens[i-1] = fmt.Sprintf("%d", int64(result))
			} else {
				tokens[i-1] = fmt.Sprintf("%g", result)
			}
			tokens = append(tokens[:i], tokens[i+2:]...)
			i -= 2
		}
	}

	// Затем обрабатываем сложение и вычитание
	for i := 1; i < len(tokens)-1; i += 2 {
		if tokens[i] == "+" || tokens[i] == "-" {
			a, err := strconv.ParseFloat(tokens[i-1], 64)
			if err != nil {
				return "Ошибка: неверный формат числа"
			}
			b, err := strconv.ParseFloat(tokens[i+1], 64)
			if err != nil {
				return "Ошибка: неверный формат числа"
			}

			var result float64
			switch tokens[i] {
			case "+":
				result = a + b
			case "-":
				result = a - b
			}

			// Заменяем обработанную часть на результат
			if result == float64(int64(result)) {
				tokens[i-1] = fmt.Sprintf("%d", int64(result))
			} else {
				tokens[i-1] = fmt.Sprintf("%g", result)
			}
			tokens = append(tokens[:i], tokens[i+2:]...)
			i -= 2
		}
	}

	if len(tokens) != 1 {
		return "Ошибка: неверный формат выражения"
	}

	return tokens[0]
}
