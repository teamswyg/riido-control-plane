package main

func mergeBackgroundErrors(channels ...<-chan error) <-chan error {
	active := make([]<-chan error, 0, len(channels))
	for _, ch := range channels {
		if ch != nil {
			active = append(active, ch)
		}
	}
	if len(active) == 0 {
		return nil
	}
	out := make(chan error, 1)
	for _, ch := range active {
		go func(ch <-chan error) {
			if err, ok := <-ch; ok {
				out <- err
			}
		}(ch)
	}
	return out
}
