package main

import (
	"bytes"
	"strings"
)

func generate(spec openAPISpec) ([]byte, error) {
	ops := flattenOperations(spec.Paths)
	if err := validateClientMetadata(spec, ops); err != nil {
		return nil, err
	}
	var b strings.Builder
	writeGeneratedHeader(&b)
	b.WriteString("import type { QueryClient, UseMutationOptions, UseQueryOptions } from '@/lib/react-query';\n\n")
	writeTypes(&b, spec.Components.Schemas)
	writeCoreRuntime(&b)
	for _, op := range ops {
		if err := writeOperation(&b, op); err != nil {
			return nil, err
		}
	}
	if err := writeFacadeInterfaces(&b, spec, ops, false); err != nil {
		return nil, err
	}
	if err := writeFacade(&b, ops); err != nil {
		return nil, err
	}
	out := bytes.TrimRight([]byte(b.String()), "\n")
	return append(out, '\n'), nil
}

func generateReact(spec openAPISpec) ([]byte, error) {
	ops := flattenOperations(spec.Paths)
	if err := validateClientMetadata(spec, ops); err != nil {
		return nil, err
	}
	var b strings.Builder
	b.WriteString("'use client';\n\n")
	writeGeneratedHeader(&b)
	b.WriteString("/* eslint-disable react-hooks/rules-of-hooks */\n\n")
	b.WriteString("import { useMemo } from 'react';\n")
	b.WriteString("import { useMutation, useQuery, type UseMutationResult, type UseQueryResult } from '@/lib/react-query';\n")
	b.WriteString("import * as core from './aiAgentClient';\n\n")
	if err := writeFacadeInterfaces(&b, spec, ops, true); err != nil {
		return nil, err
	}
	if err := writeReactFacade(&b, ops); err != nil {
		return nil, err
	}
	out := bytes.TrimRight([]byte(b.String()), "\n")
	return append(out, '\n'), nil
}
