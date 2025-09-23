package main

import "strings"

var justString string

func createHugeString(n int) string {
	return strings.Repeat("x", n)
}

func someFunc() {
	v := createHugeString(1 << 10)
	// justString = v[:100]
	// Утечка памяти из-за того, что justString по-прежнему ссылается на весь v
	// и, думаю, что сборщик мусора не будет собирать неиспользуемую память по той же причине
	justString = string([]byte(v[:100]))
}

func main() {
	someFunc()
}
