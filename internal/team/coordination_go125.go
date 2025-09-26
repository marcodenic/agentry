package team

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ExecuteParallelTasks executes multiple tasks in parallel using Go 1.25's new sync.WaitGroup.Go() method
// This demonstrates the new concurrent pattern available in Go 1.25
func (t *Team) ExecuteParallelTasks(ctx context.Context, tasks []TaskRequest) ([]*Task, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	var wg sync.WaitGroup
	results := make([]*Task, len(tasks))
	errors := make([]error, len(tasks))

	// Use Go 1.25's new WaitGroup.Go() method for cleaner concurrent patterns
	for i, taskReq := range tasks {
		i, taskReq := i, taskReq // capture loop variables

		// Go 1.25: WaitGroup.Go() method combines Add(1) + go func() pattern
		wg.Go(func() {
			agent := t.GetAgent(taskReq.AgentID)
			if agent == nil {
				errors[i] = fmt.Errorf("agent %s not found", taskReq.AgentID)
				return
			}

			task := &Task{
				ID:        uuid.New().String(),
				Type:      taskReq.Type,
				AgentID:   taskReq.AgentID,
				Input:     taskReq.Input,
				Status:    "running",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Store task immediately
			t.mutex.Lock()
			t.tasks[task.ID] = task
			t.mutex.Unlock()

			// Execute the task
			result, err := t.Call(ctx, taskReq.AgentID, taskReq.Input)
			if err != nil {
				task.Status = "failed"
				task.Result = err.Error()
				errors[i] = err
			} else {
				task.Status = "completed"
				task.Result = result
			}
			task.UpdatedAt = time.Now()

			results[i] = task
		})
	}

	// Wait for all tasks to complete
	wg.Wait()

	// Check for any errors
	var firstError error
	for _, err := range errors {
		if err != nil && firstError == nil {
			firstError = err
		}
	}

	return results, firstError
}

// TaskRequest represents a task request for parallel execution
type TaskRequest struct {
	AgentID string
	Type    string
	Input   string
}

// ExecuteWithTimeout executes tasks with a timeout using context and WaitGroup.Go()
func (t *Team) ExecuteWithTimeout(ctx context.Context, tasks []TaskRequest, timeout time.Duration) ([]*Task, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return t.ExecuteParallelTasks(ctx, tasks)
}
