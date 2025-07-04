package main

import (
"fmt"
"time"

"github.com/marcodenic/agentry/internal/config"
"github.com/marcodenic/agentry/internal/core"
"github.com/marcodenic/agentry/internal/cost"
"github.com/marcodenic/agentry/internal/memory"
"github.com/marcodenic/agentry/internal/model"
"github.com/marcodenic/agentry/internal/tool"
"github.com/marcodenic/agentry/internal/team"
)

func main() {
fmt.Printf("=== Team Agents Debug ===\n")

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

// Create team like TUI does
tm, err := team.NewTeam(ag, 10, "")
if err != nil {
fmt.Printf("Error creating team: %v\n", err)
return
}

fmt.Printf("Original agent:\n")
fmt.Printf("  ID: %s\n", ag.ID.String()[:8])
fmt.Printf("  Cost: $%.6f\n", ag.Cost.TotalCost())

// Monitor team agents like TUI does
fmt.Printf("\n=== Monitoring Team Agents ===\n")
for i := 0; i < 20; i++ {
teamAgents := tm.Agents()
fmt.Printf("Iteration %d: Found %d agents\n", i, len(teamAgents))

totalCost := 0.0
for j, agent := range teamAgents {
agentCost := 0.0
if agent.Cost != nil {
agentCost = agent.Cost.TotalCost()
}
totalCost += agentCost
fmt.Printf("  Agent %d: ID=%s, Cost=$%.6f\n", j, agent.ID.String()[:8], agentCost)
}

fmt.Printf("  Total Cost: $%.6f\n", totalCost)

time.Sleep(200 * time.Millisecond)
}
}
