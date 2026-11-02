package riidoaiserver

func (d *commentReactionDispatcher) alreadySent(id, emoji string) bool {
	if d == nil {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	value, ok := d.sent[id]
	return ok && value == emoji
}

func (d *commentReactionDispatcher) markSent(id, emoji string) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.sent) >= 2048 {
		clear(d.sent)
	}
	d.sent[id] = emoji
}
