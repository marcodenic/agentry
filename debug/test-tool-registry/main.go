package main

import (
	"fmt"
	"log"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	fmt.Println("🔧 Testing tool registry loading...")

	// Test DefaultRegistry directly
	defaultReg := tool.DefaultRegistry()
	fmt.Printf("✅ DefaultRegistry has %d tools\n", len(defaultReg))
	for name := range defaultReg {
		fmt.Printf("  - %s\n", name)
	}

	fmt.Println("\n🔧 Testing buildAgent registry...")
	
	// Load config
	cfg, err := config.Load("test-config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Show what tools are in the config
	fmt.Printf("📋 Config has %d tool manifests:\n", len(cfg.Tools))
	for _, tm := range cfg.Tools {
		fmt.Printf("  - %s\n", tm.Name)
	}

	// Build agent using the same logic as buildAgent
	tool.SetPermissions(cfg.Permissions.Tools)
	tool.SetSandboxEngine(cfg.Sandbox.Engine)
	reg := tool.Registry{}
	for _, m := range cfg.Tools {
		tl, err := tool.FromManifest(m)
		if err != nil {
			if err == tool.ErrUnknownBuiltin {
				fmt.Printf("⚠️  Skipping unknown builtin: %s\n", m.Name)
				continue
			}
			fmt.Printf("❌ Error loading tool %s: %v\n", m.Name, err)
			continue
		}
		reg[m.Name] = tl
		fmt.Printf("✅ Loaded tool: %s\n", m.Name)
	}

	fmt.Printf("\n📊 Final registry has %d tools:\n", len(reg))
	for name := range reg {
		fmt.Printf("  - %s\n", name)
	}
}
