package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := mainRun(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mainRun(args []string) error {
	fs := flag.NewFlagSet("looprefreshdispatch", flag.ContinueOnError)
	opt := options{}
	fs.StringVar(&opt.Repo, "repo", ".", "repository root or child path")
	fs.StringVar(&opt.CommandsIn, "commands-in", "", "loop refresh commands evidence")
	fs.StringVar(&opt.DispatchOut, "dispatch-out", "", "dispatch plan evidence output")
	fs.StringVar(&opt.DispatchOut, "evidence-out", "", "dispatch plan evidence output")
	fs.StringVar(&opt.CandidateOut, "candidate-out", "", "closed-loop candidate output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return run(opt)
}
