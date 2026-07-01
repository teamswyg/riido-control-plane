package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type config struct {
	BaseURL string
	Out     string
}

func main() {
	if err := run(os.Args[1:], os.Stdout, time.Now().UTC(), http.DefaultClient); err != nil {
		fmt.Fprintln(os.Stderr, "publicpageslive:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, now time.Time, client *http.Client) error {
	cfg, err := parse(args)
	if err != nil {
		return err
	}
	rec, err := collect(cfg.BaseURL, now, client)
	if err != nil {
		return err
	}
	return write(cfg.Out, stdout, rec)
}

func parse(args []string) (config, error) {
	fs := flag.NewFlagSet("publicpageslive", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var cfg config
	fs.StringVar(&cfg.BaseURL, "base-url", "", "GitHub Pages base URL")
	fs.StringVar(&cfg.Out, "out", "-", "evidence output path")
	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	if cfg.BaseURL == "" {
		return config{}, fmt.Errorf("-base-url is required")
	}
	return cfg, nil
}
