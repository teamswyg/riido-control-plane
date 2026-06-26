package main

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

type profiler struct {
	dir string
	cpu *os.File
}

func startProfiler(dir string) (profiler, error) {
	if dir == "" {
		return profiler{}, nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return profiler{}, err
	}
	cpu, err := createProfile(filepath.Join(dir, "cpu.pprof"))
	if err != nil {
		return profiler{}, err
	}
	if err := pprof.StartCPUProfile(cpu); err != nil {
		_ = cpu.Close()
		return profiler{}, err
	}
	return profiler{dir: dir, cpu: cpu}, nil
}

func (p profiler) Stop() ([]profileArtifact, error) {
	if p.dir == "" {
		return nil, nil
	}
	pprof.StopCPUProfile()
	if err := p.cpu.Close(); err != nil {
		return nil, err
	}
	if err := writeProfile(p.dir, "heap.pprof", "heap", 0); err != nil {
		return nil, err
	}
	if err := writeProfile(p.dir, "goroutine.txt", "goroutine", 1); err != nil {
		return nil, err
	}
	return profileArtifacts(p.dir)
}

func writeProfile(dir, name, profile string, debug int) error {
	if profile == "heap" {
		runtime.GC()
	}
	f, err := createProfile(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	defer f.Close()
	return pprof.Lookup(profile).WriteTo(f, debug)
}

func createProfile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
}
