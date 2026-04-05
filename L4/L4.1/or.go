package L4_1

import (
	"time"
)

func Or(channels ...<-chan interface{}) <-chan interface{} {
	orDone := make(chan interface{})

	go func() {
		defer close(orDone)

		if len(channels) == 0 {
			return
		}

		active := make([]<-chan interface{}, len(channels))
		copy(active, channels)

		for len(active) > 0 {
			for _, ch := range active {
				select {
				case <-ch:
					return
				default:
					// Канал не готов, продолжаем проверку
				}
			}
			time.Sleep(time.Microsecond * 10)
		}
	}()

	return orDone
}
