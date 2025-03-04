package orchestrator

import (
	"os"
	"strconv"
	"strings"
)

func (h *Handler) processExpression(id int, expr string) {
	tokens := strings.Fields(expr)
	if len(tokens) < 3 {
		h.mu.Lock()
		h.expressions[id] = Expression{ID: id, Status: "error", Expr: expr}
		h.mu.Unlock()
		return
	}

	// First pass: handle multiplication and division
	for i := 1; i < len(tokens)-1; i += 2 {
		if tokens[i] == "*" || tokens[i] == "/" {
			arg1, _ := strconv.ParseFloat(tokens[i-1], 64)
			arg2, _ := strconv.ParseFloat(tokens[i+1], 64)
			opTime := getOperationTime(tokens[i])
			task := Task{ID: id, Arg1: arg1, Arg2: arg2, Operation: tokens[i], OperationTime: opTime}
			h.mu.Lock()
			h.tasks[id] = task
			h.taskChan <- task
			h.mu.Unlock()

			// Replace the processed part with a placeholder
			result := performOperation(arg1, arg2, tokens[i])
			tokens[i-1] = strconv.FormatFloat(result, 'f', -1, 64)
			tokens = append(tokens[:i], tokens[i+2:]...)
			i -= 2 // Adjust index after modification
		}
	}

	// Second pass: handle addition and subtraction
	for i := 1; i < len(tokens)-1; i += 2 {
		if tokens[i] == "+" || tokens[i] == "-" {
			arg1, _ := strconv.ParseFloat(tokens[i-1], 64)
			arg2, _ := strconv.ParseFloat(tokens[i+1], 64)
			opTime := getOperationTime(tokens[i])
			task := Task{ID: id, Arg1: arg1, Arg2: arg2, Operation: tokens[i], OperationTime: opTime}
			h.mu.Lock()
			h.tasks[id] = task
			h.taskChan <- task
			h.mu.Unlock()

			// Replace the processed part with a placeholder
			result := performOperation(arg1, arg2, tokens[i])
			tokens[i-1] = strconv.FormatFloat(result, 'f', -1, 64)
			tokens = append(tokens[:i], tokens[i+2:]...)
			i -= 2 // Adjust index after modification
		}
	}
}

func performOperation(arg1, arg2 float64, op string) float64 {
	switch op {
	case "+":
		return arg1 + arg2
	case "-":
		return arg1 - arg2
	case "*":
		return arg1 * arg2
	case "/":
		return arg1 / arg2
	default:
		return 0
	}
}

func getOperationTime(op string) int {
	switch op {
	case "+":
		return atoi(os.Getenv("TIME_ADDITION_MS"))
	case "-":
		return atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	case "*":
		return atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	case "/":
		return atoi(os.Getenv("TIME_DIVISIONS_MS"))
	default:
		return 1000
	}
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
