package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"Deist_Calc/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

type HTTPServer struct {
	storage    *storage.Storage
	grpcClient CalculatorServiceClient
}

type CalculatorServiceClient interface {
	Calculate(expression, userID string) (string, error)
	GetExpressions(userID string) ([]storage.Expression, error)
}

func NewHTTPServer(storage *storage.Storage, grpcClient CalculatorServiceClient) *HTTPServer {
	return &HTTPServer{
		storage:    storage,
		grpcClient: grpcClient,
	}
}

func (s *HTTPServer) Start(port int, webDir string) error {
	// Статические файлы
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	// API endpoints
	http.HandleFunc("/api/register", s.handleRegister)
	http.HandleFunc("/api/login", s.handleLogin)
	http.HandleFunc("/api/calculate", s.handleCalculate)
	http.HandleFunc("/api/expressions", s.handleGetExpressions)
	http.HandleFunc("/api/check-auth", s.handleCheckAuth)

	log.Printf("HTTP сервер запущен на порту %d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (s *HTTPServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка обработки пароля", http.StatusInternalServerError)
		return
	}

	// Создаем пользователя
	userID, err := s.storage.CreateUser(req.Username, string(hashedPassword))
	if err != nil {
		http.Error(w, "Ошибка создания пользователя", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"user_id": userID,
	})
}

func (s *HTTPServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Получаем пользователя
	userID, hashedPassword, err := s.storage.GetUser(req.Username)
	if err != nil {
		http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"user_id": userID,
	})
}

func (s *HTTPServer) handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Expression string `json:"expression"`
		UserID     string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	result, err := s.grpcClient.Calculate(req.Expression, req.UserID)
	if err != nil {
		http.Error(w, "Ошибка вычисления", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"result": result,
	})
}

func (s *HTTPServer) handleGetExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Не указан ID пользователя", http.StatusBadRequest)
		return
	}

	expressions, err := s.grpcClient.GetExpressions(userID)
	if err != nil {
		log.Printf("Ошибка получения выражений для user_id=%s: %v", userID, err)
		http.Error(w, "Ошибка получения выражений", http.StatusInternalServerError)
		return
	}

	// Преобразуем NULL в пустую строку
	for i := range expressions {
		if expressions[i].Result == "" {
			expressions[i].Result = "вычисляется..."
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"expressions": expressions,
	})
}

func (s *HTTPServer) handleCheckAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// В реальном приложении здесь должна быть проверка сессии/токена
	// Для простоты просто возвращаем 401
	http.Error(w, "Не авторизован", http.StatusUnauthorized)
}
