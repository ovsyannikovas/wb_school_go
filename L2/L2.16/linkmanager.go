package main

import (
	"net/url"
	"path"
	"strings"
)

type LinkManager struct {
	config *Config
}

var skipExtensions = map[string]bool{
	".zip": true,
	".rar": true,
	".7z":  true,
	".tar": true,
	".gz":  true,
	".bz2": true,
	".xz":  true,
	".zst": true,

	".exe": true,
	".msi": true,
	".dmg": true,
	".pkg": true,
	".deb": true,
	".rpm": true,
	".apk": true,
	".ipa": true,
	".bin": true,
	".sh":  true,
	".bat": true,
	".cmd": true,

	".avi":  true,
	".mkv":  true,
	".mov":  true,
	".wmv":  true,
	".flv":  true,
	".m4v":  true,
	".mp4":  true,
	".mpg":  true,
	".mpeg": true,
	".webm": true,
	".3gp":  true,

	".mp3":  true,
	".wav":  true,
	".flac": true,
	".aac":  true,
	".ogg":  true,
	".m4a":  true,
	".wma":  true,

	".sql":    true,
	".db":     true,
	".sqlite": true,
	".mdb":    true,
	".accdb":  true,

	".log":    true,
	".tmp":    true,
	".temp":   true,
	".cache":  true,
	".bak":    true,
	".backup": true,

	".iso":  true,
	".img":  true,
	".dsk":  true,
	".vhd":  true,
	".vmdk": true,

	".crx":    true,
	".xpi":    true,
	".plugin": true,

	".torrent": true,
}

var skipProtocols = map[string]bool{
	"mailto:":     true,
	"tel:":        true,
	"sms:":        true,
	"fax:":        true,
	"javascript:": true,
	"data:":       true,
	"blob:":       true,
	"file:":       true,
	"ftp:":        true,
	"ftps:":       true,
	"gopher:":     true,
	"telnet:":     true,
	"ssh:":        true,
	"chrome:":     true,
	"about:":      true,
}

func NewLinkManager(config *Config) *LinkManager {
	return &LinkManager{config: config}
}

func (l *LinkManager) ShouldDownload(link string) bool {
	if link == "" {
		return false
	}

	if strings.HasPrefix(link, "#") {
		return false
	}

	linkLower := strings.ToLower(link)
	for protocol := range skipProtocols {
		if strings.HasPrefix(linkLower, protocol) {
			return false
		}
	}

	parsedURL, err := url.Parse(link)
	if err != nil {
		return false
	}

	ext := strings.ToLower(path.Ext(parsedURL.Path))
	if skipExtensions[ext] {
		return false
	}

	if strings.Contains(parsedURL.RawQuery, "download") ||
		strings.Contains(parsedURL.RawQuery, "file") ||
		strings.Contains(parsedURL.RawQuery, "attachment") {
		return false
	}

	return true
}

// IsResource проверяет, является ли ссылка ресурсом (CSS, JS, изображение и т.д.)
func (l *LinkManager) IsResource(link string) bool {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return false
	}

	ext := strings.ToLower(path.Ext(parsedURL.Path))

	return !skipExtensions[ext] && ext != ""
}

// IsHTML проверяет, является ли ссылка HTML страницей
func (l *LinkManager) IsHTML(link string) bool {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return false
	}

	ext := strings.ToLower(path.Ext(parsedURL.Path))

	return ext == "" || ext == ".html" || ext == ".htm" || ext == ".php" ||
		ext == ".asp" || ext == ".aspx" || ext == ".jsp" || ext == ".do" ||
		ext == ".action" || ext == ".cgi"
}

// NormalizeURL нормализует URL для сравнения
func (l *LinkManager) NormalizeURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	parsed.Fragment = ""

	if parsed.Path == "" {
		parsed.Path = "/"
	}

	parsed.Path = path.Clean(parsed.Path)

	return parsed.String(), nil
}

// GetFilename возвращает имя файла для сохранения
func (l *LinkManager) GetFilename(link string) string {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "index.html"
	}

	pathParts := strings.Split(parsedURL.Path, "/")
	lastPart := pathParts[len(pathParts)-1]

	if lastPart == "" {
		return "index.html"
	}

	if !strings.Contains(lastPart, ".") {
		return lastPart + ".html"
	}

	return lastPart
}

// ShouldFollow проверяет, нужно ли переходить по ссылке для дальнейшего скачивания
func (l *LinkManager) ShouldFollow(link string) bool {
	if !l.ShouldDownload(link) {
		return false
	}

	parsedURL, err := url.Parse(link)
	if err != nil {
		return false
	}

	if parsedURL.Host != "" && parsedURL.Host != l.config.Domain {
		return false
	}

	return l.IsHTML(link)
}

// IsSameDomain проверяет, принадлежит ли ссылка тому же домену
func (l *LinkManager) IsSameDomain(link string) bool {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return false
	}

	if parsedURL.Host == "" {
		return true
	}

	return parsedURL.Host == l.config.Domain
}

// GetLinkPriority возвращает приоритет ссылки (для возможной сортировки очереди)
func (l *LinkManager) GetLinkPriority(link string) int {
	if l.IsHTML(link) {
		return 1
	}
	return 2
}

// FilterLinks фильтрует список ссылок, оставляя только подходящие для скачивания
func (l *LinkManager) FilterLinks(links []string) []string {
	var filtered []string
	seen := make(map[string]bool)

	for _, link := range links {
		if !l.ShouldDownload(link) {
			continue
		}

		normalized, err := l.NormalizeURL(link)
		if err != nil {
			continue
		}

		if !seen[normalized] {
			seen[normalized] = true
			filtered = append(filtered, normalized)
		}
	}

	return filtered
}

// GetPathForSaving возвращает путь для сохранения файла
func (l *LinkManager) GetPathForSaving(link string) string {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "index.html"
	}

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

	result := append([]string{parsedURL.Host}, pathParts...)

	return strings.Join(result, "/")
}
