package main

import (
	"errors"
	"io"
	"path/filepath"
	"testing"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestRunRejectsBadCLIAndInput(t *testing.T) {
	for name, args := range map[string][]string{
		"flag":     {"-nope"},
		"contract": {},
		"missing":  {"-contract", filepath.Join(t.TempDir(), "missing.json")},
	} {
		t.Run(name, func(t *testing.T) {
			if err := run(args, io.Discard); err == nil {
				t.Fatalf("expected run error")
			}
		})
	}
}

func TestLoadContractRejectsMalformedJSON(t *testing.T) {
	for name, body := range map[string]string{
		"bad-json": "{",
		"unknown":  `{"unknown": true}`,
		"trailing": `{} {}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "contract.json")
			writeFixtureFile(t, path, body)
			if _, err := loadContract(path); err == nil {
				t.Fatalf("expected loadContract error")
			}
		})
	}
}

func TestWriteRecordReportsOutputFailures(t *testing.T) {
	if err := writeRecord("-", checkRecord{}, failingWriter{}); err == nil {
		t.Fatalf("expected stdout write error")
	}
	blocker := filepath.Join(t.TempDir(), "blocker")
	writeFixtureFile(t, blocker, "file")
	if err := writeRecord(filepath.Join(blocker, "out.json"), checkRecord{}, io.Discard); err == nil {
		t.Fatalf("expected file write error")
	}
}
