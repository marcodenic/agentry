package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
)

// TestConcurrentAgentsWithSynctest demonstrates Go 1.25's new testing/synctest package
// for testing concurrent code with virtualized time
func TestConcurrentAgentsWithSynctest(t *testing.T) {
	// Use Go 1.25's new testing/synctest for deterministic concurrent testing
	synctest.Test(t, func(t *testing.T) {
		// Create multiple mock agents
		clients := []*parallelTestClient{
			{response: "Agent 1 response", delay: 100 * time.Millisecond, t: t},
			{response: "Agent 2 response", delay: 200 * time.Millisecond, t: t},
			{response: "Agent 3 response", delay: 150 * time.Millisecond, t: t},
		}

		registry := tool.DefaultRegistry()
		agents := make([]*core.Agent, len(clients))

		for i, client := range clients {
			agents[i] = core.New(client, "mock", registry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)
			agents[i].Role = "parallel-tester"
		}

		// Create team and add agents
		tm, err := team.NewTeam(agents[0], 3, "synctest")
		if err != nil {
			t.Fatalf("Failed to create team: %v", err)
		}
		if err := tm.AddExistingAgent("agent_0", agents[0]); err != nil {
			t.Fatalf("Failed to register agent_0: %v", err)
		}

		// Add the other agents to the team
		for i := 1; i < len(agents); i++ {
			agentName := fmt.Sprintf("agent_%d", i)
			tm.Add(agentName, agents[i])
		}

		// Execute parallel tasks using Go 1.25's WaitGroup.Go()
		tasks := []team.TaskRequest{
			{AgentID: "agent_0", Type: "test", Input: "task 1"},
			{AgentID: "agent_1", Type: "test", Input: "task 2"},
			{AgentID: "agent_2", Type: "test", Input: "task 3"},
		}

		// In synctest bubble, time is virtualized
		start := time.Now()
		results, err := tm.ExecuteParallelTasks(context.Background(), tasks)
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("Parallel execution failed: %v", err)
		}

		if len(results) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(results))
		}

		// Verify all tasks completed
		for i, result := range results {
			if result == nil {
				t.Fatalf("Result %d is nil", i)
			}
			if result.Status != "completed" {
				t.Errorf("Task %d status: %s, expected: completed", i, result.Status)
			}
			if result.Result != clients[i].response {
				t.Errorf("Task %d result: %s, expected: %s", i, result.Result, clients[i].response)
			}
		}

		// In synctest, virtualized time should be minimal since goroutines run deterministically
		t.Logf("Parallel execution completed in %v (virtualized time)", elapsed)
	})
}

// TestTimeoutBehaviorWithSynctest tests timeout behavior with virtualized time
func TestTimeoutBehaviorWithSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		client := &parallelTestClient{
			response: "Never reached",
			delay:    5 * time.Second, // Longer than timeout
			t:        t,
		}

		registry := tool.DefaultRegistry()
		agent := core.New(client, "mock", registry, memory.NewInMemory(), memory.NewInMemoryVector(), nil)

		tm, err := team.NewTeam(agent, 1, "timeout-test")
		if err != nil {
			t.Fatalf("Failed to create team: %v", err)
		}

		tasks := []team.TaskRequest{
			{AgentID: agent.ID.String(), Type: "test", Input: "timeout task"},
		}

		// Execute with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		start := time.Now()
		_, err = tm.ExecuteParallelTasks(ctx, tasks)
		elapsed := time.Since(start)

		// Should timeout quickly in virtualized time
		if err == nil {
			t.Error("Expected timeout error, but got none")
		}

		t.Logf("Timeout occurred after %v (virtualized time)", elapsed)
	})
}

// TestEventStreamingWithSynctest tests event streaming with concurrent writes
func TestEventStreamingWithSynctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Test trace event streaming with concurrent writers
		events := make([]trace.Event, 0, 100)
		var mu sync.Mutex

		writer := &testTraceWriter{
			write: func(ctx context.Context, e trace.Event) {
				mu.Lock()
				events = append(events, e)
				mu.Unlock()
			},
		}

		// Simulate multiple agents writing events concurrently
		var wg sync.WaitGroup
		numAgents := 5
		eventsPerAgent := 10

		for i := 0; i < numAgents; i++ {
			i := i
			wg.Go(func() {
				for j := 0; j < eventsPerAgent; j++ {
					event := trace.Event{
						Timestamp: time.Now(),
						Type:      trace.EventStepStart,
						AgentID:   string(rune('A' + i)), // Agent A, B, C, etc.
						Data:      map[string]interface{}{"step": j},
					}
					writer.write(context.Background(), event)

					// Small delay to simulate processing
					time.Sleep(10 * time.Millisecond)
				}
			})
		}

		wg.Wait()

		// Verify all events were written
		expectedEvents := numAgents * eventsPerAgent
		if len(events) != expectedEvents {
			t.Errorf("Expected %d events, got %d", expectedEvents, len(events))
		}

		// Verify events from all agents
		agentCounts := make(map[string]int)
		for _, event := range events {
			agentCounts[event.AgentID]++
		}

		if len(agentCounts) != numAgents {
			t.Errorf("Expected events from %d agents, got %d", numAgents, len(agentCounts))
		}

		for agentID, count := range agentCounts {
			if count != eventsPerAgent {
				t.Errorf("Agent %s: expected %d events, got %d", agentID, eventsPerAgent, count)
			}
		}
	})
}

// parallelTestClient simulates an AI client with configurable delay for testing
type parallelTestClient struct {
	response string
	delay    time.Duration
	t        *testing.T
}

func (p *parallelTestClient) Clone() model.Client {
	return &parallelTestClient{response: p.response, delay: p.delay, t: p.t}
}

func (p *parallelTestClient) Stream(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (<-chan model.StreamChunk, error) {
	out := make(chan model.StreamChunk, 1)

	go func() {
		defer close(out)

		// Simulate processing time
		select {
		case <-time.After(p.delay):
			out <- model.StreamChunk{
				ContentDelta: p.response,
				Done:         true,
			}
		case <-ctx.Done():
			p.t.Logf("Client canceled due to context: %v", ctx.Err())
			return
		}
	}()

	return out, nil
}

// testTraceWriter is a simple trace writer for testing
type testTraceWriter struct {
	write func(context.Context, trace.Event)
}

func (w *testTraceWriter) Write(ctx context.Context, e trace.Event) {
	w.write(ctx, e)
}
