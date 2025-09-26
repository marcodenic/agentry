package team

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AssignTask assigns a task to a specific agent
func (t *Team) AssignTask(ctx context.Context, agentID, taskType, input string) (*Task, error) {
	agent := t.GetAgent(agentID)
	if agent == nil {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	task := &Task{
		ID:        uuid.New().String(),
		Type:      taskType,
		AgentID:   agentID,
		Input:     input,
		Status:    "assigned",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.mutex.Lock()
	t.tasks[task.ID] = task
	t.mutex.Unlock()

	// Execute the task asynchronously using the unified Call path
	go func() {
		task.Status = "running"
		task.UpdatedAt = time.Now()

		result, err := t.Call(ctx, agentID, input)
		if err != nil {
			task.Status = "failed"
			task.Result = err.Error()
		} else {
			task.Status = "completed"
			task.Result = result
		}
		task.UpdatedAt = time.Now()
	}()

	return task, nil
}

// GetTask returns a task by ID
func (t *Team) GetTask(taskID string) *Task {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.tasks[taskID]
}

// ListTasks returns all tasks
func (t *Team) ListTasks() []*Task {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	tasks := make([]*Task, 0, len(t.tasks))
	for _, task := range t.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// (inbox messaging removed)

// CoordinateTask coordinates a complex task across multiple agents
func (t *Team) CoordinateTask(ctx context.Context, description string) (*Task, error) {
	// This is a simplified version - could be enhanced with workflow logic

	// For now, assign to the first available agent
	agents := t.ListAgents()
	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents available for coordination")
	}

	// Find an agent that's ready
	var selectedAgent *Agent
	for _, agent := range agents {
		if agent.GetStatus() == "ready" {
			selectedAgent = agent
			break
		}
	}

	if selectedAgent == nil {
		return nil, fmt.Errorf("no ready agents available")
	}

	return t.AssignTask(ctx, selectedAgent.ID, "coordination", description)
}

// WaitForTask waits for a task to complete with timeout
func (t *Team) WaitForTask(ctx context.Context, taskID string, timeout time.Duration) (*Task, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		task := t.GetTask(taskID)
		if task == nil {
			return nil, fmt.Errorf("task %s not found", taskID)
		}

		if task.Status == "completed" || task.Status == "failed" {
			return task, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Continue polling
		}
	}

	return nil, fmt.Errorf("task %s did not complete within timeout", taskID)
}
