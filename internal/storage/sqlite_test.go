package storage

import (
	"os"
	"testing"
)

func setupTestDB(t *testing.T) (*Storage, func()) {
	// Создаем временную базу данных для тестов
	dbPath := "test.db"
	store, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("Ошибка создания тестовой базы данных: %v", err)
	}

	// Возвращаем функцию очистки
	cleanup := func() {
		store.db.Close()
		os.Remove(dbPath)
	}

	return store, cleanup
}

func TestCreateAndGetUser(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Тест создания пользователя
	username := "testuser"
	password := "testpass"
	userID, err := store.CreateUser(username, password)
	if err != nil {
		t.Fatalf("Ошибка создания пользователя: %v", err)
	}
	if userID == "" {
		t.Error("ID пользователя пустой")
	}

	// Тест получения пользователя
	retrievedID, retrievedPass, err := store.GetUser(username)
	if err != nil {
		t.Fatalf("Ошибка получения пользователя: %v", err)
	}
	if retrievedID != userID {
		t.Errorf("ID пользователя не совпадает: получили %s, ожидали %s", retrievedID, userID)
	}
	if retrievedPass != password {
		t.Errorf("Пароль не совпадает: получили %s, ожидали %s", retrievedPass, password)
	}
}

func TestSaveAndGetExpression(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Создаем тестового пользователя
	userID, err := store.CreateUser("testuser", "testpass")
	if err != nil {
		t.Fatalf("Ошибка создания пользователя: %v", err)
	}

	// Тест сохранения выражения
	expression := "2 + 2"
	exprID, err := store.SaveExpression(userID, expression)
	if err != nil {
		t.Fatalf("Ошибка сохранения выражения: %v", err)
	}
	if exprID == "" {
		t.Error("ID выражения пустой")
	}

	// Тест получения выражений пользователя
	expressions, err := store.GetUserExpressions(userID)
	if err != nil {
		t.Fatalf("Ошибка получения выражений: %v", err)
	}
	if len(expressions) != 1 {
		t.Errorf("Неверное количество выражений: получили %d, ожидали 1", len(expressions))
	}
	if expressions[0].Expression != expression {
		t.Errorf("Выражение не совпадает: получили %s, ожидали %s", expressions[0].Expression, expression)
	}
}

func TestUpdateExpression(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Создаем тестового пользователя и выражение
	userID, err := store.CreateUser("testuser", "testpass")
	if err != nil {
		t.Fatalf("Ошибка создания пользователя: %v", err)
	}

	exprID, err := store.SaveExpression(userID, "2 + 2")
	if err != nil {
		t.Fatalf("Ошибка сохранения выражения: %v", err)
	}

	// Тест обновления выражения
	result := "4"
	status := "completed"
	err = store.UpdateExpression(exprID, result, status)
	if err != nil {
		t.Fatalf("Ошибка обновления выражения: %v", err)
	}

	// Проверяем обновление
	expressions, err := store.GetUserExpressions(userID)
	if err != nil {
		t.Fatalf("Ошибка получения выражений: %v", err)
	}
	if len(expressions) != 1 {
		t.Fatalf("Неверное количество выражений: получили %d, ожидали 1", len(expressions))
	}
	if expressions[0].Result != result {
		t.Errorf("Результат не совпадает: получили %s, ожидали %s", expressions[0].Result, result)
	}
	if expressions[0].Status != status {
		t.Errorf("Статус не совпадает: получили %s, ожидали %s", expressions[0].Status, status)
	}
}
