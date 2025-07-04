package main

import (
"context"
"fmt"
"time"

"github.com/marcodenic/agentry/internal/config"
"github.com/marcodenic/agentry/internal/core"
"github.com/marcodenic/agentry/internal/cost"
"github.com/marcodenic/agentry/internal/memory"
"github.com/marcodenic/agentry/internal/model"
"github.com/marcodenic/agentry/internal/tool"
)

func main() {
fmt.Printf("=== Agent 0 Cost Manager Debug ===\n")

// Load config like TUI does
cfg, err := config.Load("config/test-config.yaml")
if err != nil {
fmt.Printf("Error loading config: %v\n", err)
return
}

// Create agent like buildAgent does
tool.SetPermissions(cfg.Permissions.Tools)
tool.SetSandboxEngine(cfg.Sandbox.Engine)
reg := tool.Registry{}

clients := map[string]model.Client{}
for _, m := range cfg.Models {
c, err := model.FromManifest(m)
if err != nil {
fmt.Printf("Error creating model: %v\n", err)
return
}
clients[m.Name] = c
}

var client model.Client
var modelName string
if len(cfg.Models) > 0 {
primaryModel := cfg.Models[0]
client = clients[primaryModel.Name]
if primaryModel.Options != nil && primaryModel.Options["model"] != "" {
modelName = fmt.Sprintf("%s-%s", primaryModel.Provider, primaryModel.Options["model"])
} else {
modelName = primaryModel.Name
}
} else {
client = model.NewMock()
modelName = "mock"
}

var vec memory.VectorStore
switch cfg.Vector.Type {
case "qdrant":
vec = memory.NewQdrant(cfg.Vector.URL, cfg.Vector.Collection)
case "faiss":
vec = memory.NewFaiss(cfg.Vector.URL)
default:
vec = memory.NewInMemoryVector()
}

ag := core.New(client, modelName, reg, memory.NewInMemory(), vec, nil)
ag.Cost = cost.New(0, 0.0)

fmt.Printf("Agent 0 created:\n")
fmt.Printf("  ID: %s\n", ag.ID.String()[:8])
fmt.Printf("  Model: %s\n", modelName)
fmt.Printf("  Cost Manager: %v\n", ag.Cost != nil)

// Test cost before any operations
fmt.Printf("\n=== Initial Cost Test ===\n")
initialCost := ag.Cost.TotalCost()
fmt.Printf("Initial cost: $%.6f\n", initialCost)

// Test multiple calls to see if it's stable
for i := 0; i < 10; i++ {
cost := ag.Cost.TotalCost()
if cost != initialCost {
fmt.Printf("COST CHANGED! Was $%.6f, now $%.6f\n", initialCost, cost)
}
fmt.Printf("Call %d: $%.6f\n", i, cost)
time.Sleep(100 * time.Millisecond)
}

// Test with a prompt
fmt.Printf("\n=== Testing with Prompt ===\n")
response, err := ag.Run(context.Background(), "Say hello")
if err != nil {
fmt.Printf("Error running prompt: %v\n", err)
} else {
fmt.Printf("Response length: %d chars\n", len(response))
}

afterCost := ag.Cost.TotalCost()
fmt.Printf("After prompt cost: $%.6f\n", afterCost)

// Test stability after prompt
for i := 0; i < 20; i++ {
cost := ag.Cost.TotalCost()
if cost != afterCost {
fmt.Printf("COST CHANGED! Was $%.6f, now $%.6f\n", afterCost, cost)
afterCost = cost
}
fmt.Printf("Post-prompt call %d: $%.6f\n", i, cost)
time.Sleep(200 * time.Millisecond)
}
}
