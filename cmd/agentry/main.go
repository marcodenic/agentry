package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/eval"
	"github.com/marcodenic/agentry/internal/server"
	"github.com/marcodenic/agentry/internal/tui"
	"github.com/spf13/cobra"
)

func main() {
	env.Load()

	var cfgFile string

	rootCmd := &cobra.Command{Use: "agentry", Short: "Agentry CLI"}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to .agentry.yaml")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "dev",
		Short: "Run REPL using example config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("examples/.agentry.yaml")
			if err != nil {
				return err
			}
			ag, err := buildAgent(cfg)
			if err != nil {
				return err
			}
			sc := bufio.NewScanner(os.Stdin)
			fmt.Println("Agentry REPL – Ctrl-D to quit")
			for {
				fmt.Print("> ")
				if !sc.Scan() {
					break
				}
				line := sc.Text()
				out, err := ag.Run(context.Background(), line)
				if err != nil {
					fmt.Println("ERR:", err)
					continue
				}
				fmt.Println(out)
			}
			return nil
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgFile == "" {
				return fmt.Errorf("need --config")
			}
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			ag, err := buildAgent(cfg)
			if err != nil {
				return err
			}
			agents := map[string]*core.Agent{"default": ag}
			server.Serve(agents)
			return nil
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "eval",
		Short: "Run evaluation suite",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgFile == "" {
				return fmt.Errorf("need --config")
			}
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			key := os.Getenv("OPENAI_KEY")
			if key != "" {
				for i, m := range cfg.Models {
					if m.Name == "openai" {
						if m.Options == nil {
							m.Options = map[string]string{}
						}
						cfg.Models[i].Options["key"] = key
					}
				}
			}
			ag, err := buildAgent(cfg)
			if err != nil {
				return err
			}
			suite := "tests/eval_suite.json"
			if key != "" {
				suite = "tests/openai_eval_suite.json"
			}
			eval.Run(nil, ag, suite)
			return nil
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "tui",
		Short: "Launch interactive terminal UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgFile == "" {
				return fmt.Errorf("need --config")
			}
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			ag, err := buildAgent(cfg)
			if err != nil {
				return err
			}
			p := tea.NewProgram(tui.New(ag))
			return p.Start()
		},
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
