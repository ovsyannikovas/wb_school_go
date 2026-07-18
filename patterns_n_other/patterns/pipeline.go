package main

import (
	"fmt"
)

// stage - функция-этап pipeline
type Stage func(in <-chan int) <-chan int

// Pipeline строит цепочку обработки
func Pipeline(stages ...Stage) Stage {
	return func(in <-chan int) <-chan int {
		current := in
		for _, stage := range stages {
			current = stage(current)
		}
		return current
	}
}

// Создаем этапы обработки
func multiplyBy(factor int) Stage {
	return func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for val := range in {
				out <- val * factor
			}
		}()
		return out
	}
}

func add(value int) Stage {
	return func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for val := range in {
				out <- val + value
			}
		}()
		return out
	}
}

func main() {
	// Создаем пайплайн: *2 -> +10 -> *3
	pipeline := Pipeline(
		multiplyBy(2),
		add(10),
		multiplyBy(3),
	)

	// Входной канал
	input := make(chan int)

	// Запускаем пайплайн
	output := pipeline(input)

	// Отправляем данные
	go func() {
		for i := 1; i <= 5; i++ {
			input <- i
		}
		close(input)
	}()

	// Читаем результаты
	for result := range output {
		fmt.Printf("%d -> ((%d*2)+10)*3 = %d\n", result/6, result/6, result)
	}
}
