package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	url := flag.String("url", "", "URL сайта для зеркалирования")
	depth := flag.Int("depth", 3, "Глубина рекурсии")
	parallel := flag.Int("parallel", 5, "Количество параллельных загрузок")
	output := flag.String("output", "mirror", "Директория для сохранения")
	timeout := flag.Duration("timeout", 30*time.Second, "Таймаут HTTP запросов")

	flag.Parse()

	if err := os.MkdirAll(*output, 0755); err != nil {
		log.Fatalf("Ошибка создания директории %s: %v", *output, err)
	}

	config := &Config{
		BaseURL:   *url,
		MaxDepth:  *depth,
		Parallel:  *parallel,
		OutputDir: *output,
		Timeout:   *timeout,
		Visited:   make(map[string]bool),
		Domain:    extractDomain(*url),
	}

	downloader := NewDownloader(config)
	fmt.Printf("Начинаем зеркалирование %s\n", *url)
	fmt.Printf("Параметры: глубина=%d, параллельность=%d, директория=%s\n",
		*depth, *parallel, *output)
	fmt.Println(strings.Repeat("─", 50))

	if err := downloader.Start(); err != nil {
		log.Fatalf("Ошибка при зеркалировании: %v", err)
	}

	fmt.Println(strings.Repeat("─", 50))
	fmt.Println("Зеркалирование успешно завершено!")
}

func extractDomain(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "ftp://")
	url = strings.TrimPrefix(url, "file://")

	parts := strings.Split(url, ":")
	host := parts[0]

	parts = strings.Split(host, "/")
	return parts[0]
}
