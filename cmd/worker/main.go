package main

import (
	"fmt"
	"os"

	"github.com/marcodenic/agentry/internal/taskqueue"
)

func main() {
	q, err := taskqueue.NewQueue(natsURL(), "agentry.tasks")
	if err != nil {
		panic("NATS unavailable: " + err.Error())
	}
	defer q.Close()

	// agents := make(map[string]*core.Agent) // TODO: load/configure agents as needed

	fmt.Println("Worker listening for tasks...")
	_, err = q.Subscribe(func(task taskqueue.Task) {
		if task.Type == "invoke" {
			payload, ok := task.Payload.(map[string]interface{})
			if !ok {
				fmt.Println("bad payload")
				return
			}
			agentID, _ := payload["agent_id"].(string)
			input, _ := payload["input"].(string)
			// TODO: lookup agent, run task, handle result
			fmt.Printf("[Worker] Invoking agent %s with input: %s\n", agentID, input)
			// ag := agents[agentID]
			// out, err := ag.Run(context.Background(), input)
			// ... handle output ...
		}
	})
	if err != nil {
		panic(err)
	}
	select {} // block forever
}

func natsURL() string {
	if u := os.Getenv("NATS_URL"); u != "" {
		return u
	}
	return "nats://localhost:4222"
}
