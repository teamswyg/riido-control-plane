package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	if err := runMain(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "aiagentthreadsnapshot:", err)
		os.Exit(1)
	}
}

func runMain(args []string) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}
	report, err := capture(context.Background(), cfg)
	if err != nil {
		return err
	}
	return writeReport(cfg.OutputPath, report)
}

type config struct {
	BaseURL        string
	WorkspaceID    string
	TaskID         string
	ConversationID string
	TokenEnv       string
	OutputPath     string
	SSEWindow      time.Duration
}

func parseConfig(args []string) (config, error) {
	var cfg config
	fs := flag.NewFlagSet("aiagentthreadsnapshot", flag.ContinueOnError)
	fs.StringVar(&cfg.BaseURL, "base-url", "https://staging.ai-api.riido.io", "control-plane API base URL")
	fs.StringVar(&cfg.WorkspaceID, "workspace-id", "", "workspace id")
	fs.StringVar(&cfg.TaskID, "task-id", "", "task id")
	fs.StringVar(&cfg.ConversationID, "conversation-id", "", "optional conversation id to highlight")
	fs.StringVar(&cfg.TokenEnv, "token-env", "RIIDO_PRIVATE_JWT", "environment variable containing bearer token")
	fs.StringVar(&cfg.OutputPath, "output", "", "output JSON path")
	fs.DurationVar(&cfg.SSEWindow, "sse-window", 5*time.Second, "SSE capture window")
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	if cfg.WorkspaceID == "" || cfg.TaskID == "" || cfg.OutputPath == "" {
		return cfg, fmt.Errorf("workspace-id, task-id, and output are required")
	}
	return cfg, nil
}
