// ainativepreflight validates production task configuration without AWS writes.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func validate(body []byte, name string) error {
	var envelope struct {
		Task       *task       `json:"taskDefinition"`
		Containers []container `json:"containerDefinitions"`
	}
	if json.Unmarshal(body, &envelope) != nil {
		return errors.New("invalid task definition JSON")
	}
	containers := envelope.Containers
	if envelope.Task != nil {
		containers = envelope.Task.Containers
	}
	var chosen []container
	for _, c := range containers {
		if c.Name == name {
			chosen = append(chosen, c)
		}
	}
	if name == "" || len(chosen) != 1 {
		return errors.New("configured container must occur exactly once")
	}
	if len(chosen[0].EnvironmentFiles) != 0 {
		return errors.New("native AI requires explicit task settings; environment files cannot be verified")
	}
	env, secrets := map[string]string{}, map[string]string{}
	for _, e := range chosen[0].Environment {
		if _, ok := env[e.Name]; ok {
			return errors.New("duplicate environment entry")
		}
		env[e.Name] = e.Value
	}
	for _, e := range chosen[0].Secrets {
		if _, ok := secrets[e.Name]; ok {
			return errors.New("duplicate secret entry")
		}
		if _, ok := env[e.Name]; ok {
			return errors.New("environment and secret names overlap")
		}
		secrets[e.Name] = e.ValueFrom
	}
	for _, k := range []string{"RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_ID", "RIIDO_AI_SERVER_TASK_CONTEXT_TEAM_ID", "RIIDO_AI_SERVER_TASK_CONTEXT_WORKSPACE_API_KEY"} {
		_, secretPresent := secrets[k]
		if strings.TrimSpace(env[k]) != "" || secretPresent {
			return fmt.Errorf("%s must be absent for requester-private native context", k)
		}
	}
	for k, v := range map[string]string{"RIIDO_AI_SERVER_TASK_CONTEXT_LOCATOR_VERSION": "2", "RIIDO_AI_SERVER_ASSIGNMENT_READ_SCOPE": "true"} {
		if env[k] != v {
			return fmt.Errorf("%s must be %s", k, v)
		}
	}
	for _, k := range keyNames {
		if strings.TrimSpace(secrets[k]) == "" {
			return fmt.Errorf("%s requires a secret reference", k)
		}
	}
	for k, path := range urlPaths {
		u, err := url.Parse(env[k])
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" || strings.TrimSuffix(u.Path, "/") != path || nonProduction.MatchString(strings.ToLower(u.Hostname())) {
			return fmt.Errorf("%s requires a production HTTPS endpoint with the expected path", k)
		}
	}
	return nil
}
