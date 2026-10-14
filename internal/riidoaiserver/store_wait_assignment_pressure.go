package riidoaiserver

import "time"

const (
	riidoPollElapsedMsKey = "riido.poll.elapsed_ms"
	riidoPollHoldMsKey    = "riido.poll.hold_ms"
	riidoPollTickMsKey    = "riido.poll.tick_ms"
	riidoPollWaitedKey    = "riido.poll.waited"
)

func pollDurationMs(duration time.Duration) int64 {
	return int64(duration / time.Millisecond)
}
