package riidoaiserver

import "context"

func (s *Store) Metrics(ctx context.Context) (MetricsSnapshot, error) {
	reply := make(chan metricsResult, 1)
	if err := s.send(ctx, metricsCmd{reply: reply}); err != nil {
		return MetricsSnapshot{}, err
	}
	select {
	case res := <-reply:
		return res.snapshot, res.err
	case <-ctx.Done():
		return MetricsSnapshot{}, ctx.Err()
	}
}
