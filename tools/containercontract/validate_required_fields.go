package main

import (
	"fmt"
	"strings"
)

type requiredContractField struct {
	Name  string
	Value func(imageContract) string
}

var requiredContractFields = []requiredContractField{
	{"service", func(c imageContract) string { return c.Service }},
	{"dockerfile", func(c imageContract) string { return c.Dockerfile }},
	{"build.build_arg.name", func(c imageContract) string { return c.Build.BuildArg.Name }},
	{"build.build_arg.default", func(c imageContract) string { return c.Build.BuildArg.Default }},
	{"build.stage_name", func(c imageContract) string { return c.Build.StageName }},
	{"build.workdir", func(c imageContract) string { return c.Build.Workdir }},
	{"build.cgo_enabled", func(c imageContract) string { return c.Build.CGOEnabled }},
	{"build.module_download", func(c imageContract) string { return c.Build.ModuleDownload.Command }},
	{"build.go_build.package", func(c imageContract) string { return c.Build.GoBuild.Package }},
	{"build.go_build.output", func(c imageContract) string { return c.Build.GoBuild.Output }},
	{"final.base_image", func(c imageContract) string { return c.Final.BaseImage }},
	{"final.copy_from", func(c imageContract) string { return c.Final.CopyFrom }},
	{"final.copy_source", func(c imageContract) string { return c.Final.CopySource }},
	{"final.binary", func(c imageContract) string { return c.Final.Binary }},
	{"final.user", func(c imageContract) string { return c.Final.User }},
}

func requireContractFields(contract imageContract) error {
	for _, field := range requiredContractFields {
		if strings.TrimSpace(field.Value(contract)) == "" {
			return fmt.Errorf("%s is required", field.Name)
		}
	}
	return nil
}
