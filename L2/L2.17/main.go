package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "таймаут подключения")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Использование: %s [--timeout duration] host port\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Пример: %s --timeout 5s localhost 8080\n", os.Args[0])
		os.Exit(1)
	}

	address := args[0] + ":" + args[1]
	fmt.Fprintf(os.Stderr, "Подключение к %s (таймаут: %v)...\n", address, *timeout)

	conn, err := net.DialTimeout("tcp", address, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка подключения: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Fprintln(os.Stderr, "Подключено успешно. Для выхода нажмите Ctrl+C или Ctrl+D")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer cancel()

		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			if err != io.EOF && !isClosedError(err) {
				fmt.Fprintf(os.Stderr, "\nОшибка чтения из сокета: %v\n", err)
			}
		}
		fmt.Fprintln(os.Stderr, "\nСоединение закрыто сервером")
	}()

	go func() {
		defer wg.Done()
		defer cancel()

		_, err := io.Copy(conn, os.Stdin)
		if err != nil {
			if err != io.EOF && !isClosedError(err) {
				fmt.Fprintf(os.Stderr, "Ошибка записи в сокет: %v\n", err)
			}
		}
		fmt.Fprintln(os.Stderr, "Получен сигнал завершения от пользователя")
	}()

	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nПолучен сигнал прерывания, завершаем работу...")
		cancel()
	}()

	<-ctx.Done()

	go func() {
		time.Sleep(2 * time.Second)
		fmt.Fprintln(os.Stderr, "Таймаут при завершении, принудительный выход...")
		os.Exit(1)
	}()

	wg.Wait()
	fmt.Fprintln(os.Stderr, "Программа завершена")
}

func isClosedError(err error) bool {
	if err == nil {
		return false
	}
	opErr, ok := err.(*net.OpError)
	if ok {
		return opErr.Err.Error() == "use of closed network connection"
	}
	return false
}
