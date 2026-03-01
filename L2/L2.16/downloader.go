package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Downloader struct {
	config    *Config
	client    *http.Client
	taskQueue chan *Task
	wg        sync.WaitGroup
	parser    *HTMLParser
	linkMgr   *LinkManager
	stats     *Stats
}

type Task struct {
	URL   string
	Depth int
}

type Stats struct {
	mu         sync.Mutex
	downloaded int
	failed     int
	skipped    int
	bytesTotal int64
	startTime  time.Time
}

func NewDownloader(config *Config) *Downloader {
	return &Downloader{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
		taskQueue: make(chan *Task, 1000),
		parser:    NewHTMLParser(config),
		linkMgr:   NewLinkManager(config),
		stats:     &Stats{startTime: time.Now()},
	}
}

func (d *Downloader) Start() error {
	fmt.Printf("Начинаем зеркалирование %s\n", d.config.BaseURL)
	fmt.Printf("Глубина: %d, Параллельных загрузок: %d\n", d.config.MaxDepth, d.config.Parallel)
	fmt.Printf("Директория сохранения: %s\n", d.config.OutputDir)
	fmt.Println("----------------------------------------")

	for i := 0; i < d.config.Parallel; i++ {
		d.wg.Add(1)
		go d.worker(i)
	}

	d.taskQueue <- &Task{
		URL:   d.config.BaseURL,
		Depth: 0,
	}

	go d.monitorProgress()

	d.wg.Wait()
	close(d.taskQueue)

	d.printStats()
	return nil
}

func (d *Downloader) worker(id int) {
	defer d.wg.Done()

	for task := range d.taskQueue {
		if d.config.IsVisited(task.URL) {
			d.stats.incrementSkipped()
			continue
		}

		fmt.Printf("[Воркер %d] Загрузка: %s (глубина %d)\n", id, task.URL, task.Depth)

		if err := d.downloadResource(task.URL, task.Depth); err != nil {
			d.stats.incrementFailed()
			fmt.Printf("[Воркер %d] Ошибка %s: %v\n", id, task.URL, err)
		} else {
			d.stats.incrementDownloaded()
		}
	}
}

func (d *Downloader) downloadResource(resourceURL string, depth int) error {
	parsedURL, err := url.Parse(resourceURL)
	if err != nil {
		return err
	}

	if parsedURL.Host == "" {
		base, _ := url.Parse(d.config.BaseURL)
		parsedURL = base.ResolveReference(parsedURL)
		resourceURL = parsedURL.String()
	}

	if !d.config.IsSameDomain(resourceURL) {
		return fmt.Errorf("внешний домен: %s", parsedURL.Hostname())
	}

	if !d.linkMgr.ShouldDownload(resourceURL) {
		return fmt.Errorf("ресурс пропущен (расширение или протокол)")
	}

	req, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoWget/1.0; +http://example.com)")
	req.Header.Set("Accept", "*/*")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP статус %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	localPath := d.getLocalPath(parsedURL)

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(localPath, bodyBytes, 0644); err != nil {
		return err
	}

	written := int64(len(bodyBytes))
	d.stats.addBytes(written)

	d.config.MarkVisited(resourceURL)
	fmt.Printf("Сохранено: %s (%d KB)\n", localPath, written/1024)

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") && depth < d.config.MaxDepth {
		links, err := d.parser.ExtractLinks(bytes.NewReader(bodyBytes), resourceURL)
		if err != nil {
			fmt.Printf("Ошибка парсинга HTML %s: %v\n", resourceURL, err)
			return nil
		}

		for _, link := range links {
			if link != "" && !d.config.IsVisited(link) && d.linkMgr.ShouldDownload(link) {
				select {
				case d.taskQueue <- &Task{
					URL:   link,
					Depth: depth + 1,
				}:
				default:
					fmt.Printf("Очередь переполнена, пропускаем: %s\n", link)
				}
			}
		}
		fmt.Printf("Найдено ссылок: %d на странице %s\n", len(links), resourceURL)
	}

	return nil
}

func (d *Downloader) getLocalPath(parsedURL *url.URL) string {
	pathParts := strings.Split(parsedURL.Path, "/")

	if len(pathParts) == 0 || pathParts[len(pathParts)-1] == "" {
		if len(pathParts) > 0 && pathParts[len(pathParts)-1] == "" {
			pathParts = pathParts[:len(pathParts)-1]
		}
		pathParts = append(pathParts, "index.html")
	}

	lastPart := pathParts[len(pathParts)-1]
	if !strings.Contains(lastPart, ".") {
		pathParts[len(pathParts)-1] = lastPart + ".html"
	}

	// Формирование локального пути
	localPath := d.config.OutputDir
	localPath = filepath.Join(localPath, parsedURL.Host)
	localPath = filepath.Join(localPath, filepath.Join(pathParts...))

	localPath = filepath.Clean(localPath)

	return localPath
}

func (d *Downloader) monitorProgress() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		d.stats.mu.Lock()
		fmt.Printf("\rПрогресс: %d загружено, %d ошибок, %d пропущено, всего %.2f MB, в очереди: %d",
			d.stats.downloaded,
			d.stats.failed,
			d.stats.skipped,
			float64(d.stats.bytesTotal)/1024/1024,
			len(d.taskQueue))
		d.stats.mu.Unlock()
	}
}

func (d *Downloader) printStats() {
	d.stats.mu.Lock()
	defer d.stats.mu.Unlock()

	elapsed := time.Since(d.stats.startTime)
	fmt.Println("\n\nСтатистика зеркалирования:")
	fmt.Println("----------------------------------------")
	fmt.Printf("Успешно загружено: %d\n", d.stats.downloaded)
	fmt.Printf("Ошибок: %d\n", d.stats.failed)
	fmt.Printf("Пропущено (уже загружено/внешние): %d\n", d.stats.skipped)
	fmt.Printf("Всего данных: %.2f MB\n", float64(d.stats.bytesTotal)/1024/1024)
	fmt.Printf("Время выполнения: %v\n", elapsed.Round(time.Second))
	fmt.Printf("Директория: %s\n", d.config.OutputDir)

	if d.stats.downloaded > 0 {
		avgSize := float64(d.stats.bytesTotal) / float64(d.stats.downloaded) / 1024
		fmt.Printf("Средний размер файла: %.2f KB\n", avgSize)
	}
	fmt.Println("----------------------------------------")
}

func (s *Stats) incrementDownloaded() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.downloaded++
}

func (s *Stats) incrementFailed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failed++
}

func (s *Stats) incrementSkipped() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.skipped++
}

func (s *Stats) addBytes(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bytesTotal += bytes
}
