// containercontract validates the riido_ai_server Dockerfile against the
// public executable container image contract.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

const contractSchemaVersion = "riido-container-image-contract.v1"
const checkSchemaVersion = "riido-container-image-contract-check.v1"

type imageContract struct {
	SchemaVersion           string        `json:"schema_version"`
	Service                 string        `json:"service"`
	Dockerfile              string        `json:"dockerfile"`
	FargateTaskDefinitionIR string        `json:"fargate_task_definition_ir,omitempty"`
	Build                   buildContract `json:"build"`
	Final                   finalContract `json:"final"`
}

type buildContract struct {
	BuildArg   buildArgContract `json:"build_arg"`
	StageName  string           `json:"stage_name"`
	Workdir    string           `json:"workdir"`
	CGOEnabled string           `json:"cgo_enabled"`
	GoBuild    goBuildContract  `json:"go_build"`
}

type buildArgContract struct {
	Name    string `json:"name"`
	Default string `json:"default"`
}

type goBuildContract struct {
	Package  string   `json:"package"`
	Output   string   `json:"output"`
	Trimpath bool     `json:"trimpath"`
	LDFlags  []string `json:"ldflags"`
}

type finalContract struct {
	BaseImage      string                 `json:"base_image"`
	CopyFrom       string                 `json:"copy_from"`
	CopySource     string                 `json:"copy_source"`
	Binary         string                 `json:"binary"`
	RequiredCopies []requiredCopyContract `json:"required_copies,omitempty"`
	ExposedPorts   []int                  `json:"exposed_ports"`
	Env            map[string]string      `json:"env"`
	User           string                 `json:"user"`
	Entrypoint     []string               `json:"entrypoint"`
}

type requiredCopyContract struct {
	From        string `json:"from"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type dockerfile struct {
	Args   map[string]string
	Stages []stage
}

type stage struct {
	Base       string
	Alias      string
	Workdir    string
	Env        map[string]string
	Runs       []string
	Copies     []copyInstruction
	Exposes    []int
	User       string
	Entrypoint []string
}

type copyInstruction struct {
	From string
	Src  string
	Dst  string
}

type taskDefinitionIR struct {
	RuntimePlatform struct {
		OperatingSystemFamily string `json:"operatingSystemFamily"`
	} `json:"runtime_platform"`
	Container struct {
		PortMappings []struct {
			ContainerPort int `json:"containerPort"`
		} `json:"portMappings"`
		Environment []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"environment"`
	} `json:"container"`
}

type checkRecord struct {
	SchemaVersion           string   `json:"schema_version"`
	Service                 string   `json:"service"`
	Dockerfile              string   `json:"dockerfile"`
	FargateTaskDefinitionIR string   `json:"fargate_task_definition_ir,omitempty"`
	BuildStage              string   `json:"build_stage"`
	FinalBaseImage          string   `json:"final_base_image"`
	FinalUser               string   `json:"final_user"`
	Entrypoint              []string `json:"entrypoint"`
	ExposedPorts            []int    `json:"exposed_ports"`
	ChecksTotal             int      `json:"checks_total"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "containercontract:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("containercontract", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	contractPath := fs.String("contract", "", "path to riido-container-image-contract.v1")
	outPath := fs.String("out", "", "optional path to write riido-container-image-contract-check.v1 JSON, or - for stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*contractPath) == "" {
		return errors.New("-contract is required")
	}
	contract, err := loadContract(*contractPath)
	if err != nil {
		return err
	}
	record, err := verifyContract(contract)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*outPath) == "" {
		return nil
	}
	body, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if *outPath == "-" {
		_, err := stdout.Write(body)
		return err
	}
	return os.WriteFile(*outPath, body, 0o644)
}

func loadContract(path string) (imageContract, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return imageContract{}, err
	}
	var contract imageContract
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&contract); err != nil {
		return imageContract{}, fmt.Errorf("decode container contract: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return imageContract{}, errors.New("decode container contract: trailing data")
	}
	if err := validateContract(contract); err != nil {
		return imageContract{}, err
	}
	return contract, nil
}

func validateContract(contract imageContract) error {
	if contract.SchemaVersion != contractSchemaVersion {
		return fmt.Errorf("schema_version must be %s", contractSchemaVersion)
	}
	required := map[string]string{
		"service":                 contract.Service,
		"dockerfile":              contract.Dockerfile,
		"build.build_arg.name":    contract.Build.BuildArg.Name,
		"build.build_arg.default": contract.Build.BuildArg.Default,
		"build.stage_name":        contract.Build.StageName,
		"build.workdir":           contract.Build.Workdir,
		"build.cgo_enabled":       contract.Build.CGOEnabled,
		"build.go_build.package":  contract.Build.GoBuild.Package,
		"build.go_build.output":   contract.Build.GoBuild.Output,
		"final.base_image":        contract.Final.BaseImage,
		"final.copy_from":         contract.Final.CopyFrom,
		"final.copy_source":       contract.Final.CopySource,
		"final.binary":            contract.Final.Binary,
		"final.user":              contract.Final.User,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", field)
		}
	}
	if len(contract.Final.ExposedPorts) == 0 {
		return errors.New("final.exposed_ports is required")
	}
	if len(contract.Final.Entrypoint) == 0 {
		return errors.New("final.entrypoint is required")
	}
	if len(contract.Final.Env) == 0 {
		return errors.New("final.env is required")
	}
	for i, copy := range contract.Final.RequiredCopies {
		if strings.TrimSpace(copy.From) == "" {
			return fmt.Errorf("final.required_copies[%d].from is required", i)
		}
		if strings.TrimSpace(copy.Source) == "" {
			return fmt.Errorf("final.required_copies[%d].source is required", i)
		}
		if strings.TrimSpace(copy.Destination) == "" {
			return fmt.Errorf("final.required_copies[%d].destination is required", i)
		}
	}
	return nil
}

func verifyContract(contract imageContract) (checkRecord, error) {
	parsed, err := parseDockerfile(contract.Dockerfile)
	if err != nil {
		return checkRecord{}, err
	}
	checks := 0
	require := func(ok bool, format string, args ...any) error {
		checks++
		if !ok {
			return fmt.Errorf(format, args...)
		}
		return nil
	}
	if err := require(parsed.Args[contract.Build.BuildArg.Name] == contract.Build.BuildArg.Default,
		"ARG %s default = %q, want %q", contract.Build.BuildArg.Name, parsed.Args[contract.Build.BuildArg.Name], contract.Build.BuildArg.Default); err != nil {
		return checkRecord{}, err
	}
	buildStage := parsed.stageByAlias(contract.Build.StageName)
	if buildStage == nil {
		return checkRecord{}, fmt.Errorf("build stage %q not found", contract.Build.StageName)
	}
	if err := require(buildStage.Base == "${"+contract.Build.BuildArg.Name+"}", "build stage base = %q, want ${%s}", buildStage.Base, contract.Build.BuildArg.Name); err != nil {
		return checkRecord{}, err
	}
	if err := require(buildStage.Workdir == contract.Build.Workdir, "build WORKDIR = %q, want %q", buildStage.Workdir, contract.Build.Workdir); err != nil {
		return checkRecord{}, err
	}
	if err := require(buildStage.Env["CGO_ENABLED"] == contract.Build.CGOEnabled, "CGO_ENABLED = %q, want %q", buildStage.Env["CGO_ENABLED"], contract.Build.CGOEnabled); err != nil {
		return checkRecord{}, err
	}
	if err := require(hasGoBuildRun(buildStage.Runs, contract.Build.GoBuild), "go build command does not satisfy container contract"); err != nil {
		return checkRecord{}, err
	}
	finalStage := parsed.finalStage()
	if finalStage == nil {
		return checkRecord{}, errors.New("final stage not found")
	}
	if err := require(finalStage.Base == contract.Final.BaseImage, "final base image = %q, want %q", finalStage.Base, contract.Final.BaseImage); err != nil {
		return checkRecord{}, err
	}
	if err := require(finalStage.User == contract.Final.User, "final USER = %q, want %q", finalStage.User, contract.Final.User); err != nil {
		return checkRecord{}, err
	}
	if err := require(reflect.DeepEqual(finalStage.Entrypoint, contract.Final.Entrypoint), "ENTRYPOINT = %v, want %v", finalStage.Entrypoint, contract.Final.Entrypoint); err != nil {
		return checkRecord{}, err
	}
	if err := require(intSetEqual(finalStage.Exposes, contract.Final.ExposedPorts), "EXPOSE = %v, want %v", finalStage.Exposes, contract.Final.ExposedPorts); err != nil {
		return checkRecord{}, err
	}
	for key, want := range contract.Final.Env {
		if err := require(finalStage.Env[key] == want, "final ENV %s = %q, want %q", key, finalStage.Env[key], want); err != nil {
			return checkRecord{}, err
		}
	}
	if err := require(hasCopy(finalStage.Copies, contract.Final.CopyFrom, contract.Final.CopySource, contract.Final.Binary), "final COPY from %s %s -> %s not found", contract.Final.CopyFrom, contract.Final.CopySource, contract.Final.Binary); err != nil {
		return checkRecord{}, err
	}
	for _, copy := range contract.Final.RequiredCopies {
		if err := require(hasCopy(finalStage.Copies, copy.From, copy.Source, copy.Destination), "final required COPY from %s %s -> %s not found", copy.From, copy.Source, copy.Destination); err != nil {
			return checkRecord{}, err
		}
	}
	if err := require(contract.Final.User != "0" && !strings.HasPrefix(contract.Final.User, "0:"), "final user must be non-root"); err != nil {
		return checkRecord{}, err
	}
	if strings.TrimSpace(contract.FargateTaskDefinitionIR) != "" {
		if err := verifyTaskDefinitionIR(contract); err != nil {
			return checkRecord{}, err
		}
		checks += 3
	}
	return checkRecord{
		SchemaVersion:           checkSchemaVersion,
		Service:                 contract.Service,
		Dockerfile:              filepath.ToSlash(contract.Dockerfile),
		FargateTaskDefinitionIR: filepath.ToSlash(contract.FargateTaskDefinitionIR),
		BuildStage:              contract.Build.StageName,
		FinalBaseImage:          finalStage.Base,
		FinalUser:               finalStage.User,
		Entrypoint:              append([]string(nil), finalStage.Entrypoint...),
		ExposedPorts:            sortedInts(finalStage.Exposes),
		ChecksTotal:             checks,
	}, nil
}

func verifyTaskDefinitionIR(contract imageContract) error {
	body, err := os.ReadFile(contract.FargateTaskDefinitionIR)
	if err != nil {
		return err
	}
	var task taskDefinitionIR
	dec := json.NewDecoder(strings.NewReader(string(body)))
	if err := dec.Decode(&task); err != nil {
		return fmt.Errorf("decode fargate task definition IR: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("decode fargate task definition IR: trailing data")
	}
	if task.RuntimePlatform.OperatingSystemFamily != "LINUX" {
		return fmt.Errorf("runtime platform OS = %q, want LINUX", task.RuntimePlatform.OperatingSystemFamily)
	}
	taskPorts := map[int]bool{}
	for _, mapping := range task.Container.PortMappings {
		taskPorts[mapping.ContainerPort] = true
	}
	for _, port := range contract.Final.ExposedPorts {
		if !taskPorts[port] {
			return fmt.Errorf("task definition does not expose container port %d", port)
		}
	}
	taskEnv := map[string]string{}
	for _, env := range task.Container.Environment {
		taskEnv[env.Name] = env.Value
	}
	for key, want := range contract.Final.Env {
		if taskEnv[key] != want {
			return fmt.Errorf("task definition env %s = %q, want %q", key, taskEnv[key], want)
		}
	}
	return nil
}

func parseDockerfile(path string) (dockerfile, error) {
	file, err := os.Open(path)
	if err != nil {
		return dockerfile{}, err
	}
	defer file.Close()
	var parsed dockerfile
	parsed.Args = map[string]string{}
	var current *stage
	for _, logical := range logicalDockerfileLines(file) {
		instruction, rest, ok := splitInstruction(logical)
		if !ok {
			continue
		}
		switch instruction {
		case "ARG":
			name, value := splitKeyValue(rest)
			parsed.Args[name] = value
		case "FROM":
			base, alias := parseFrom(rest)
			parsed.Stages = append(parsed.Stages, stage{Base: base, Alias: alias, Env: map[string]string{}})
			current = &parsed.Stages[len(parsed.Stages)-1]
		case "WORKDIR":
			if current != nil {
				current.Workdir = strings.TrimSpace(rest)
			}
		case "ENV":
			if current != nil {
				name, value := splitKeyValue(rest)
				current.Env[name] = value
			}
		case "RUN":
			if current != nil {
				current.Runs = append(current.Runs, strings.TrimSpace(rest))
			}
		case "COPY":
			if current != nil {
				current.Copies = append(current.Copies, parseCopy(rest))
			}
		case "EXPOSE":
			if current != nil {
				current.Exposes = append(current.Exposes, parseExposedPorts(rest)...)
			}
		case "USER":
			if current != nil {
				current.User = strings.TrimSpace(rest)
			}
		case "ENTRYPOINT":
			if current != nil {
				current.Entrypoint = parseExecJSON(rest)
			}
		}
	}
	return parsed, nil
}

func logicalDockerfileLines(r io.Reader) []string {
	var lines []string
	scanner := bufio.NewScanner(r)
	var current strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		continued := strings.HasSuffix(line, "\\")
		line = strings.TrimSpace(strings.TrimSuffix(line, "\\"))
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(line)
		if continued {
			continue
		}
		lines = append(lines, current.String())
		current.Reset()
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

func splitInstruction(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", "", false
	}
	instruction := strings.ToUpper(parts[0])
	return instruction, strings.TrimSpace(strings.TrimPrefix(line, parts[0])), true
}

func splitKeyValue(rest string) (string, string) {
	rest = strings.TrimSpace(rest)
	if before, after, ok := strings.Cut(rest, "="); ok {
		return strings.TrimSpace(before), strings.Trim(strings.TrimSpace(after), `"`)
	}
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func parseFrom(rest string) (string, string) {
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return "", ""
	}
	base := parts[0]
	if len(parts) >= 3 && strings.EqualFold(parts[1], "AS") {
		return base, parts[2]
	}
	return base, ""
}

func parseCopy(rest string) copyInstruction {
	parts := strings.Fields(rest)
	var out copyInstruction
	var positional []string
	for _, part := range parts {
		if strings.HasPrefix(part, "--from=") {
			out.From = strings.TrimPrefix(part, "--from=")
			continue
		}
		if strings.HasPrefix(part, "--") {
			continue
		}
		positional = append(positional, part)
	}
	if len(positional) >= 2 {
		out.Src = positional[len(positional)-2]
		out.Dst = positional[len(positional)-1]
	}
	return out
}

func parseExposedPorts(rest string) []int {
	var out []int
	for _, part := range strings.Fields(rest) {
		portPart, _, _ := strings.Cut(part, "/")
		port, err := strconv.Atoi(portPart)
		if err == nil {
			out = append(out, port)
		}
	}
	return out
}

func parseExecJSON(rest string) []string {
	var out []string
	if err := json.Unmarshal([]byte(rest), &out); err == nil {
		return out
	}
	return nil
}

func (d dockerfile) stageByAlias(alias string) *stage {
	for i := range d.Stages {
		if d.Stages[i].Alias == alias {
			return &d.Stages[i]
		}
	}
	return nil
}

func (d dockerfile) finalStage() *stage {
	if len(d.Stages) == 0 {
		return nil
	}
	return &d.Stages[len(d.Stages)-1]
}

func hasGoBuildRun(runs []string, want goBuildContract) bool {
	for _, run := range runs {
		if !strings.Contains(run, "go build") {
			continue
		}
		if !strings.Contains(run, "-o "+want.Output) || !strings.Contains(run, want.Package) {
			continue
		}
		if want.Trimpath && !strings.Contains(run, "-trimpath") {
			continue
		}
		matchesFlags := true
		for _, flag := range want.LDFlags {
			if !strings.Contains(run, flag) {
				matchesFlags = false
				break
			}
		}
		if matchesFlags {
			return true
		}
	}
	return false
}

func hasCopy(copies []copyInstruction, from, src, dst string) bool {
	for _, copy := range copies {
		if copy.From == from && copy.Src == src && copy.Dst == dst {
			return true
		}
	}
	return false
}

func intSetEqual(a, b []int) bool {
	a = sortedInts(a)
	b = sortedInts(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sortedInts(in []int) []int {
	out := append([]int(nil), in...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
