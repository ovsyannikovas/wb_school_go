package main

// Подсказка:
// generic
// Tee разветвляет входной канал на n выходных каналов
// outputChs, одна горутина, конвертация в <-

func Tee[T any](inputCh <-chan T, n int) []<-chan T {
	outputChs := make([]chan T, n)
	for i := 0; i < n; i++ {
		outputChs[i] = make(chan T, 10)
	}

	// Запускаем горутину для разветвления
	go func() {
		defer func() {
			// Закрываем все выходные каналы
			for _, ch := range outputChs {
				close(ch)
			}
		}()

		for value := range inputCh {
			// Отправляем значение во все каналы
			for i := 0; i < n; i++ {
				outputChs[i] <- value
			}
		}
	}()

	// Конвертируем []chan T в []<-chan T
	resultChs := make([]<-chan T, n)
	for i := 0; i < n; i++ {
		resultChs[i] = outputChs[i]
	}

	return resultChs
}
