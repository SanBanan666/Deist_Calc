package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	pb "Deist_Calc/internal/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverProcess *exec.Cmd
	httpPort      = 8080
	grpcPort      = 50051
)

func setupTestServer(t *testing.T) func() {
	// Запускаем сервер в отдельном процессе
	serverProcess = exec.Command("go", "run", "./cmd/server/main.go",
		"-http-port", fmt.Sprintf("%d", httpPort),
		"-grpc-port", fmt.Sprintf("%d", grpcPort),
		"-db", "test_integration.db",
	)
	serverProcess.Stdout = os.Stdout
	serverProcess.Stderr = os.Stderr

	if err := serverProcess.Start(); err != nil {
		t.Fatalf("Ошибка запуска сервера: %v", err)
	}

	// Ждем запуска сервера
	time.Sleep(2 * time.Second)

	// Возвращаем функцию очистки
	return func() {
		serverProcess.Process.Kill()
		serverProcess.Wait()
		os.Remove("test_integration.db")
	}
}

func TestUserRegistrationAndLogin(t *testing.T) {
	cleanup := setupTestServer(t)
	defer cleanup()

	// Тест регистрации
	registerData := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	jsonData, _ := json.Marshal(registerData)

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/api/register", httpPort),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Ошибка регистрации: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Неверный статус регистрации: %d", resp.StatusCode)
	}

	var registerResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}
	userID := registerResp["user_id"]
	if userID == "" {
		t.Error("ID пользователя пустой")
	}

	// Тест входа
	loginData := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	jsonData, _ = json.Marshal(loginData)

	resp, err = http.Post(
		fmt.Sprintf("http://localhost:%d/api/login", httpPort),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Ошибка входа: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Неверный статус входа: %d", resp.StatusCode)
	}

	var loginResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}
	if loginResp["user_id"] != userID {
		t.Errorf("ID пользователя не совпадает: получили %s, ожидали %s", loginResp["user_id"], userID)
	}
}

func TestCalculatorWorkflow(t *testing.T) {
	cleanup := setupTestServer(t)
	defer cleanup()

	// Регистрируем пользователя
	registerData := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	jsonData, _ := json.Marshal(registerData)

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/api/register", httpPort),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Ошибка регистрации: %v", err)
	}

	var registerResp map[string]string
	json.NewDecoder(resp.Body).Decode(&registerResp)
	userID := registerResp["user_id"]

	// Подключаемся к gRPC серверу
	conn, err := grpc.Dial(
		fmt.Sprintf("localhost:%d", grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Ошибка подключения к gRPC серверу: %v", err)
	}
	defer conn.Close()

	client := pb.NewCalculatorServiceClient(conn)

	// Тест вычисления выражения
	calcResp, err := client.Calculate(context.Background(), &pb.CalculateRequest{
		Expression: "2 + 2",
		UserId:     userID,
	})
	if err != nil {
		t.Fatalf("Ошибка вычисления: %v", err)
	}
	if calcResp.Result == "" {
		t.Error("Результат вычисления пустой")
	}

	// Тест получения выражений
	exprResp, err := client.GetExpressions(context.Background(), &pb.GetExpressionsRequest{
		UserId: userID,
	})
	if err != nil {
		t.Fatalf("Ошибка получения выражений: %v", err)
	}
	if len(exprResp.Expressions) == 0 {
		t.Error("Список выражений пуст")
	}
	found := false
	for _, expr := range exprResp.Expressions {
		if expr.Expression == "2 + 2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Выражение не найдено в списке")
	}
}
