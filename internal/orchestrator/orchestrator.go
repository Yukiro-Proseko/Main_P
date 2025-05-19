package orchestrator

import (
	"log"
	"paral/internal/config"
	"paral/pkg/calc"

	"github.com/google/uuid"
)

func (r *Repository) AddExpression(expression *config.Expression) {
	log.Printf("adding expression: ID=%s, Tasks: %v\n", expression.ID, expression.Tasks)

	r.mu.Lock()
	r.expressions[expression.ID] = expression
	r.mu.Unlock()

	for _, task := range expression.Tasks {
		log.Printf("Adding Task: ID=%s, Arg1=%s, Arg2=%s, Operation=%s\n", task.ID, task.Arg1, task.Arg2, task.Operation)
		r.tasks[task.ID] = task
	}
}

func (r *Repository) GetExpressionByID(expressionID string) (*config.Expression, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	expression, exists := r.expressions[expressionID]

	return expression, exists
}
func (r *Repository) GetAllExpressions() []*config.Expression {
	r.mu.RLock()
	defer r.mu.RUnlock()

	expressions := make([]*config.Expression, 0, len(r.expressions))

	for _, expression := range r.expressions {
		expressions = append(expressions, expression)
	}

	return expressions
}

func (r *Repository) GetTaskByID(taskID string) (*config.Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[taskID]

	return task, exists
}

func (r *Repository) GetPendingTask() (*config.Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, task := range r.tasks {
		if task.Status == "pending" && r.allDependenciesCompleted(task.Dependencies) {
			return task, true
		}
	}
	return nil, false
}

func (r *Repository) allDependenciesCompleted(dependencies []string) bool {
	for _, depID := range dependencies {
		depTask, exists := r.tasks[depID]
		if !exists || depTask.Status != "completed" {
			return false
		}
	}
	return true
}

func (r *Repository) UpdateTaskStatus(taskID string, status string, result float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[taskID]
	if !exists {
		log.Printf("", taskID)
		return
	}

	task.Status = status
	if status == "completed" {
		task.Result = result
	} else if status == "error" {
		for _, expr := range r.expressions {
			for _, t := range expr.Tasks {
				if t.ID == taskID {
					expr.Status = "error"
					break
				}
			}
		}
	}

	if task.ExpressionID != "" {
		expr, exists := r.expressions[task.ExpressionID]
		if exists {
			allCompleted := true
			for _, t := range expr.Tasks {
				storedTask, found := r.tasks[t.ID]
				if !found || storedTask.Status != "completed" {
					allCompleted = false
					break
				}
			}
			if allCompleted {
				lastTaskID := expr.Tasks[len(expr.Tasks)-1].ID
				if lastTask, found := r.tasks[lastTaskID]; found {
					expr.Status = "completed"
					expr.Result = lastTask.Result
					log.Printf("Expression %s updated to completed with result: %f", task.ExpressionID, lastTask.Result)
				} else {
					log.Printf("Last task %s not found for expression %s", lastTaskID, task.ExpressionID)
				}
			} else {
				expr.Status = "pending"
				log.Printf("Expression %s remains pending", task.ExpressionID)
			}
		}
	}
}

func (r *Repository) UpdateExpressionStatus(expressionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	expression, exists := r.expressions[expressionID]
	if !exists {
		log.Printf("Expression %s not found in repository", expressionID)
		return
	}

	allCompleted := true

	for _, task := range expression.Tasks {
		storedTask, found := r.tasks[task.ID]
		if !found || storedTask.Status != "completed" {
			allCompleted = false
			break
		}
	}

	if allCompleted {

		lastTaskID := expression.Tasks[len(expression.Tasks)-1].ID
		if lastTask, found := r.tasks[lastTaskID]; found {
			expression.Status = "completed"
			expression.Result = lastTask.Result
			log.Printf("Expression %s updated to completed with result: %f", expressionID, lastTask.Result)
		} else {
			log.Printf("Last task %s not found for expression %s", lastTaskID, expressionID)
		}
	} else {
		expression.Status = "pending"
		log.Printf("Expression %s remains pending", expressionID)
	}

	r.expressions[expressionID] = expression
}

func (a *Application) AddExpression(expression string) (string, error) {
	expressionID := uuid.New().String()

	tasks, err := calc.ParseExpression(expression, expressionID)
	if err != nil {
		return "", err
	}

	a.repository.AddExpression(&config.Expression{
		ID:     expressionID,
		Status: "pending",
		Tasks:  tasks,
	})

	return expressionID, nil
}

func (a *Application) GetExpressionByID(expressionID string) (*config.Expression, bool) {
	return a.repository.GetExpressionByID(expressionID)
}

func (a *Application) GetAllExpressions() []*config.Expression {
	return a.repository.GetAllExpressions()
}

func (a *Application) GetPendingTask() (*config.Task, bool) {
	return a.repository.GetPendingTask()
}

func (a *Application) UpdateTaskStatus(taskID string, status string, result float64) {
	a.repository.UpdateTaskStatus(taskID, status, result)
}
