package main

type metricsCaptureWriter struct {
	ch chan string
}

func (w metricsCaptureWriter) Write(p []byte) (int, error) {
	w.ch <- string(p)
	return len(p), nil
}
