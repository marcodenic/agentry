package main

import (
	"fmt"
	"log"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	// Create a basic agent (simulating Agent 0)
	ag := &core.Agent{
		Tools: tool.DefaultRegistry(),
	}

	// Test NewTeamContext - should create empty team
	team, err := converse.NewTeamContext(ag)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Initial team agents: %d\n", len(team.Agents()))
	fmt.Printf("Initial team names: %v\n", team.Names())

	// Verify that we can add agents dynamically
	newAgent, name := team.AddAgent("TestAgent")
	fmt.Printf("Added agent: %s (ID: %s)\n", name, newAgent.ID)
	fmt.Printf("Team agents after adding: %d\n", len(team.Agents()))
	fmt.Printf("Team names after adding: %v\n", team.Names())
}
