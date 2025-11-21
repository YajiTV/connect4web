package httphandler

// notify broadcasts a tick to all subscribers of the room
func notify(rm *Room) {
	if rm == nil {
		return
	}
	for ch := range rm.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// subscribe registers a buffered channel and returns an unsubscribe function
func subscribe(rm *Room) (chan struct{}, func()) {
	ch := make(chan struct{}, 1)

	roomsMu.Lock()
	if rm.subs == nil {
		rm.subs = make(map[chan struct{}]struct{})
	}
	rm.subs[ch] = struct{}{}
	roomsMu.Unlock()

	unsub := func() {
		roomsMu.Lock()
		delete(rm.subs, ch)
		roomsMu.Unlock()
	}
	return ch, unsub
}
