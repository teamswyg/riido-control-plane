package riidoaiserver

func (s *Store) Close() {
	select {
	case <-s.done:
		return
	default:
	}
	reply := make(chan struct{})
	select {
	case s.commands <- closeCmd{reply: reply}:
		<-reply
	case <-s.done:
	}
}
