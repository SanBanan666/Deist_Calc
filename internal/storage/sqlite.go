package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия базы данных: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	st := &Storage{db: db}
	_ = st.ForceClearAllPending() // очищаем все старые зависшие задачи при запуске
	_ = st.CleanupStaleTasks()    // и дополнительно чистим старые задачи

	return st, nil
}

func createTables(db *sql.DB) error {
	// Создаем таблицу пользователей
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы пользователей: %v", err)
	}

	// Создаем таблицу выражений
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS expressions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			expression TEXT NOT NULL,
			result TEXT,
			status TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы выражений: %v", err)
	}

	return nil
}

func (s *Storage) CreateUser(username, passwordHash string) (string, error) {
	userID := generateUUID()
	_, err := s.db.Exec(
		"INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)",
		userID, username, passwordHash,
	)
	if err != nil {
		return "", fmt.Errorf("ошибка создания пользователя: %v", err)
	}
	return userID, nil
}

func (s *Storage) GetUser(username string) (string, string, error) {
	var userID, passwordHash string
	err := s.db.QueryRow(
		"SELECT id, password_hash FROM users WHERE username = ?",
		username,
	).Scan(&userID, &passwordHash)
	if err != nil {
		return "", "", fmt.Errorf("ошибка получения пользователя: %v", err)
	}
	return userID, passwordHash, nil
}

func (s *Storage) SaveExpression(userID, expression string) (string, error) {
	fmt.Printf("Сохраняем выражение: userID=%s, expression=%s\n", userID, expression)

	// Проверяем, есть ли уже такое выражение в статусе pending
	var existingID string
	err := s.db.QueryRow(
		"SELECT id FROM expressions WHERE user_id = ? AND expression = ? AND status = 'pending'",
		userID, expression,
	).Scan(&existingID)
	if err == nil {
		fmt.Printf("Найдено существующее выражение: id=%s\n", existingID)
		return existingID, nil
	} else if err != sql.ErrNoRows {
		return "", fmt.Errorf("ошибка проверки существующего выражения: %v", err)
	}

	// Очищаем старые записи (оставляем только последние 10)
	_, err = s.db.Exec(`
		DELETE FROM expressions 
		WHERE user_id = ? 
		AND id NOT IN (
			SELECT id FROM expressions 
			WHERE user_id = ? 
			ORDER BY created_at DESC 
			LIMIT 10
		)`,
		userID, userID,
	)
	if err != nil {
		return "", fmt.Errorf("ошибка очистки старых записей: %v", err)
	}

	exprID := generateUUID()
	_, err = s.db.Exec(
		"INSERT INTO expressions (id, user_id, expression, status) VALUES (?, ?, ?, ?)",
		exprID, userID, expression, "pending",
	)
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения выражения: %v", err)
	}
	fmt.Printf("Сохранено новое выражение: id=%s\n", exprID)
	return exprID, nil
}

func (s *Storage) UpdateExpression(id, result, status string) error {
	fmt.Printf("Обновляем выражение: id=%s, result=%s, status=%s\n", id, result, status)

	// Проверим, существует ли выражение
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM expressions WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка проверки существования выражения: %v", err)
	}
	if !exists {
		return fmt.Errorf("выражение не найдено")
	}

	_, err = s.db.Exec(
		"UPDATE expressions SET result = ?, status = ? WHERE id = ?",
		result, status, id,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления выражения: %v", err)
	}
	fmt.Printf("Выражение успешно обновлено\n")
	return nil
}

func (s *Storage) GetUserExpressions(userID string) ([]Expression, error) {
	rows, err := s.db.Query(
		"SELECT id, user_id, expression, result, status, created_at FROM expressions WHERE user_id = ? ORDER BY id DESC LIMIT 20",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения выражений: %v", err)
	}
	defer rows.Close()

	var expressions []Expression
	for rows.Next() {
		var expr Expression
		var createdAt time.Time
		var result sql.NullString
		err := rows.Scan(&expr.ID, &expr.UserID, &expr.Expression, &result, &expr.Status, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования выражения: %v", err)
		}
		if result.Valid {
			expr.Result = result.String
		} else {
			expr.Result = ""
		}
		expr.CreatedAt = createdAt.Format(time.RFC3339)
		expressions = append(expressions, expr)
	}
	return expressions, nil
}

func (s *Storage) GetPendingExpression() (*Expression, error) {
	fmt.Println("Получаем pending выражение из базы данных")

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции: %v", err)
	}

	var expr Expression
	var createdAt time.Time
	var result sql.NullString
	row := tx.QueryRow(
		"SELECT id, user_id, expression, result, status, created_at FROM expressions WHERE status = 'pending' ORDER BY created_at ASC LIMIT 1",
	)
	err = row.Scan(&expr.ID, &expr.UserID, &expr.Expression, &result, &expr.Status, &createdAt)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			fmt.Println("Нет выражений в статусе pending")
			return nil, fmt.Errorf("нет доступных задач")
		}
		return nil, fmt.Errorf("ошибка получения выражения: %v", err)
	}
	if result.Valid {
		expr.Result = result.String
	} else {
		expr.Result = ""
	}
	expr.CreatedAt = createdAt.Format(time.RFC3339)

	// Ставим статус processing
	_, err = tx.Exec("UPDATE expressions SET status = 'processing' WHERE id = ?", expr.ID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("ошибка обновления статуса: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("ошибка коммита транзакции: %v", err)
	}

	fmt.Printf("Получено выражение: id=%s, expression=%s, status=%s\n", expr.ID, expr.Expression, expr.Status)
	return &expr, nil
}

func (s *Storage) CleanupStaleTasks() error {
	_, err := s.db.Exec(`
		UPDATE expressions
		SET status = 'error', result = 'Истекло время ожидания'
		WHERE (status = 'pending' OR status = 'processing')
		AND created_at < datetime('now', '-10 minutes')
	`)
	if err != nil {
		return fmt.Errorf("ошибка очистки зависших задач: %v", err)
	}
	return nil
}

func (s *Storage) ForceClearAllPending() error {
	_, err := s.db.Exec(`UPDATE expressions SET status = 'error', result = 'Сброшено вручную' WHERE status = 'pending' OR status = 'processing'`)
	return err
}

type Expression struct {
	ID         string
	UserID     string
	Expression string
	Result     string
	Status     string
	CreatedAt  string
}

func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
