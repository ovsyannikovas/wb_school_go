package sync

// Подсказка:
// 1 поле, буфер 1

type ChannelMutex struct {
	ch chan struct{}
}

func NewChannelMutex() *ChannelMutex {
	return &ChannelMutex{ch: make(chan struct{}, 1)}
}

func (m *ChannelMutex) Lock() {
	m.ch <- struct{}{} // блокируется, если канал занят
}

func (m *ChannelMutex) Unlock() {
	<-m.ch
}
