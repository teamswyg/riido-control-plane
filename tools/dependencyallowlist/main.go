package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const schemaVersion = "riido-go-dependency-allowlist.v1"

type contract struct {
	SchemaVersion        string          `json:"schema_version"`
	Service              string          `json:"service"`
	Policy               string          `json:"policy"`
	AllowedDirectModules []allowedModule `json:"allowed_direct_modules"`
}

type allowedModule struct {
	Path     string `json:"path"`
	Category string `json:"category"`
	Owner    string `json:"owner"`
	Reason   string `json:"reason"`
}

type goModule struct {
	Path     string
	Version  string
	Main     bool
	Indirect bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "dependencyallowlist:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("dependencyallowlist", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	contractPath := fs.String("contract", "dependency_allowlist.riido.json", "dependency allowlist contract JSON path")
	moduleDir := fs.String("module", ".", "Go module directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	allowlist, err := loadContract(*contractPath)
	if err != nil {
		return err
	}
	modules, err := listModules(*moduleDir)
	if err != nil {
		return err
	}
	report, err := verifyModules(allowlist, modules)
	if err != nil {
		return err
	}
	fmt.Println(report)
	return nil
}

func loadContract(path string) (contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return contract{}, fmt.Errorf("read contract: %w", err)
	}
	var c contract
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&c); err != nil {
		return contract{}, fmt.Errorf("decode contract: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return contract{}, errors.New("decode contract: trailing JSON data")
	} else if !errors.Is(err, io.EOF) {
		return contract{}, fmt.Errorf("decode contract trailer: %w", err)
	}
	if c.SchemaVersion != schemaVersion {
		return contract{}, fmt.Errorf("schema_version = %q, want %q", c.SchemaVersion, schemaVersion)
	}
	if strings.TrimSpace(c.Service) == "" {
		return contract{}, errors.New("service is required")
	}
	if strings.TrimSpace(c.Policy) == "" {
		return contract{}, errors.New("policy is required")
	}
	seen := map[string]struct{}{}
	for i, module := range c.AllowedDirectModules {
		if strings.TrimSpace(module.Path) == "" {
			return contract{}, fmt.Errorf("allowed_direct_modules[%d].path is required", i)
		}
		if strings.TrimSpace(module.Category) == "" || strings.TrimSpace(module.Owner) == "" || strings.TrimSpace(module.Reason) == "" {
			return contract{}, fmt.Errorf("allowed_direct_modules[%d] must include category, owner, and reason", i)
		}
		if _, ok := seen[module.Path]; ok {
			return contract{}, fmt.Errorf("duplicate allowed direct module %q", module.Path)
		}
		seen[module.Path] = struct{}{}
	}
	return c, nil
}

func listModules(dir string) ([]goModule, error) {
	cmd := exec.Command("go", "list", "-m", "-json", "all")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("go list -m -json all: %w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("go list -m -json all: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	var modules []goModule
	for {
		var module goModule
		if err := decoder.Decode(&module); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decode go list module: %w", err)
		}
		modules = append(modules, module)
	}
	return modules, nil
}

func verifyModules(c contract, modules []goModule) (string, error) {
	allowed := map[string]allowedModule{}
	for _, module := range c.AllowedDirectModules {
		allowed[module.Path] = module
	}
	var direct []goModule
	var disallowed []goModule
	used := map[string]struct{}{}
	for _, module := range modules {
		if module.Main {
			continue
		}
		if module.Indirect {
			continue
		}
		direct = append(direct, module)
		if _, ok := allowed[module.Path]; !ok {
			disallowed = append(disallowed, module)
			continue
		}
		used[module.Path] = struct{}{}
	}
	sort.Slice(direct, func(i, j int) bool { return direct[i].Path < direct[j].Path })
	sort.Slice(disallowed, func(i, j int) bool { return disallowed[i].Path < disallowed[j].Path })
	if len(disallowed) > 0 {
		return "", fmt.Errorf("disallowed direct Go dependencies:\n%s", formatModules(disallowed))
	}
	var unused []string
	for _, module := range c.AllowedDirectModules {
		if _, ok := used[module.Path]; !ok {
			unused = append(unused, module.Path)
		}
	}
	sort.Strings(unused)
	if len(unused) > 0 {
		return "", fmt.Errorf("unused direct dependency allowlist entries:\n%s", strings.Join(unused, "\n"))
	}
	return fmt.Sprintf("verified %d approved direct Go dependencies for %s", len(direct), c.Service), nil
}

func formatModules(modules []goModule) string {
	var lines []string
	for _, module := range modules {
		line := module.Path
		if module.Version != "" {
			line += " " + module.Version
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
