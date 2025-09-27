package team

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/core"
)

func TestAssignTaskLifecycle(t *testing.T) {
	t.Setenv("AGENTRY_TUI_MODE", "1")
	parent := newTestParentAgent()
	teamInstance, err := NewTeam(parent, 3, "squad")
	if err != nil {
		t.Fatalf("NewTeam: %v", err)
	}

	agent, err := teamInstance.SpawnAgent(context.Background(), "coder", "coder")
	if err != nil {
		t.Fatalf("SpawnAgent: %v", err)
	}

	originalRunAgent := runAgentFn
	defer func() { runAgentFn = originalRunAgent }()

	ready := make(chan struct{})
	resume := make(chan struct{})

	runAgentFn = func(ctx context.Context, ag *core.Agent, input, name string, peers []string) (string, error) {
		close(ready)
		<-resume
		return "done", nil
	}

	task, err := teamInstance.AssignTask(context.Background(), agent.Name, "analysis", "Inspect logs")
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}

	<-ready

	snapshot := teamInstance.GetTask(task.ID)
	if snapshot == nil || snapshot.Status != "running" {
		t.Fatalf("expected task running, got %#v", snapshot)
	}

	close(resume)

	completed, err := teamInstance.WaitForTask(context.Background(), task.ID, time.Second)
	if err != nil {
		t.Fatalf("WaitForTask: %v", err)
	}
	if completed.Status != "completed" {
		t.Fatalf("expected completed status, got %s", completed.Status)
	}
	if completed.Result != "done" {
		t.Fatalf("expected recorded result, got %s", completed.Result)
	}
}

func TestExecuteParallelTasksRunsAll(t *testing.T) {
	t.Setenv("AGENTRY_TUI_MODE", "1")
	parent := newTestParentAgent()
	teamInstance, err := NewTeam(parent, 3, "crew")
	if err != nil {
		t.Fatalf("NewTeam: %v", err)
	}

	coder, err := teamInstance.SpawnAgent(context.Background(), "coder", "coder")
	if err != nil {
		t.Fatalf("SpawnAgent coder: %v", err)
	}
	reviewer, err := teamInstance.SpawnAgent(context.Background(), "reviewer", "reviewer")
	if err != nil {
		t.Fatalf("SpawnAgent reviewer: %v", err)
	}

	originalRunAgent := runAgentFn
	defer func() { runAgentFn = originalRunAgent }()

	var mu sync.Mutex
	responses := map[string]string{
		coder.Name:    "code done",
		reviewer.Name: "review ok",
	}

	runAgentFn = func(ctx context.Context, ag *core.Agent, input, name string, peers []string) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		return responses[name], nil
	}

	tasks, err := teamInstance.ExecuteParallelTasks(context.Background(), []TaskRequest{
		{AgentID: coder.Name, Type: "build", Input: "Implement"},
		{AgentID: reviewer.Name, Type: "review", Input: "Review"},
	})
	if err != nil {
		t.Fatalf("ExecuteParallelTasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected two tasks, got %d", len(tasks))
	}

	for _, task := range tasks {
		if task.Status != "completed" {
			t.Fatalf("expected task completed, got %s", task.Status)
		}
		if task.Result == "" {
			t.Fatalf("expected task result recorded")
		}
	}

	snapshots := teamInstance.ListTasks()
	if len(snapshots) != 2 {
		t.Fatalf("expected stored tasks, got %d", len(snapshots))
	}
	for _, snapshot := range snapshots {
		if snapshot.Status != "completed" {
			t.Fatalf("expected stored task completed, got %s", snapshot.Status)
		}
	}
}
