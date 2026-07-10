package riidoaiserver

import "time"

func storeOpenRetryDelay(attempt int) time.Duration {
	switch attempt {
	case 0:
		return 200 * time.Millisecond
	case 1:
		return 400 * time.Millisecond
	case 2:
		return 800 * time.Millisecond
	case 3:
		return 1600 * time.Millisecond
	case 4:
		return 3200 * time.Millisecond
	default:
		return 10 * time.Second
	}
}
