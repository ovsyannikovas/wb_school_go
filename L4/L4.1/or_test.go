package L4_1

import (
	"testing"
	"time"
)

func TestOr_Basic(t *testing.T) {
	ch1 := make(chan interface{})
	ch2 := make(chan interface{})

	go func() {
		time.Sleep(100 * time.Millisecond)
		close(ch2)
	}()

	start := time.Now()
	<-Or(ch1, ch2)
	elapsed := time.Since(start)

	if elapsed < 90*time.Millisecond {
		t.Errorf("Слишком быстро: %v", elapsed)
	}
	if elapsed > 150*time.Millisecond {
		t.Errorf("Слишком медленно: %v", elapsed)
	}
}

func TestOr_Empty(t *testing.T) {
	result := Or()

	select {
	case <-result:
		// ОК - канал закрыт
	case <-time.After(10 * time.Millisecond):
		t.Error("Канал не закрылся")
	}
}

func TestOr_Single(t *testing.T) {
	ch := make(chan interface{})

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(ch)
	}()

	start := time.Now()
	<-Or(ch)
	elapsed := time.Since(start)

	if elapsed < 40*time.Millisecond {
		t.Errorf("Слишком быстро: %v", elapsed)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("Слишком медленно: %v", elapsed)
	}
}

func TestOr_AlreadyClosed(t *testing.T) {
	ch1 := make(chan interface{})
	ch2 := make(chan interface{})

	close(ch1)

	start := time.Now()
	<-Or(ch1, ch2)
	elapsed := time.Since(start)

	if elapsed > 1*time.Millisecond {
		t.Errorf("Слишком медленно для закрытого канала: %v", elapsed)
	}
}

func TestOr_ValueNotClose(t *testing.T) {
	ch := make(chan interface{})

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch <- "hello" // отправляем значение, НЕ закрываем
	}()

	start := time.Now()
	<-Or(ch)
	elapsed := time.Since(start)

	if elapsed < 40*time.Millisecond {
		t.Errorf("Слишком быстро: %v", elapsed)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("Слишком медленно: %v", elapsed)
	}
}

func TestOr_MultipleChannels(t *testing.T) {
	ch1 := make(chan interface{})
	ch2 := make(chan interface{})
	ch3 := make(chan interface{})
	ch4 := make(chan interface{})

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(ch3)
	}()

	start := time.Now()
	<-Or(ch1, ch2, ch3, ch4)
	elapsed := time.Since(start)

	if elapsed < 40*time.Millisecond || elapsed > 100*time.Millisecond {
		t.Errorf("Ожидалось ~50ms, получено %v", elapsed)
	}
}
