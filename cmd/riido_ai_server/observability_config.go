package main

func observabilityConfigFromEnv() (tracingRuntimeConfig, string, error) {
	tracing, err := tracingConfigFromEnv()
	if err != nil {
		return tracingRuntimeConfig{}, "", err
	}
	pprofAddr, err := parsePprofAddr(getenvDefault(envPprofAddr, ""))
	if err != nil {
		return tracingRuntimeConfig{}, "", err
	}
	return tracing, pprofAddr, nil
}
