package main

import (
	"L4_2/grepper"
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		pattern     string
		file        string
		workers     string
		concurrency int
		showHelp    bool
	)

	flag.StringVar(&pattern, "p", "", "Pattern to search (regexp supported)")
	flag.StringVar(&file, "f", "", "File to search in")
	flag.StringVar(&workers, "workers", "", "Comma-separated list of worker addresses (host:port)")
	flag.IntVar(&concurrency, "c", 4, "Concurrency level per worker")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.Parse()

	if showHelp || pattern == "" || file == "" {
		printHelp()
		return
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Printf("Error: File '%s' does not exist\n", file)
		os.Exit(1)
	}

	grepper.Run(file, pattern, workers, concurrency)
}

func printHelp() {
	fmt.Printf(`Distributed Grep Utility
    
Usage: mygrep [options]

Options:
  -p <pattern>     Pattern to search (supports regexp)
  -f <file>        File to search in
  -workers <list>    Comma-separated list of worker workers (host:port)
  -c <num>         Concurrency level per worker (default: 4)
  -h               Show this help

Examples:
  # Local mode (fallback)
  mygrep -p "error" -f app.log
  
  # Distributed mode with 3 workers
  mygrep -p "error" -f app.log -workers "localhost:8081,localhost:8082,localhost:8083"
  
  # High concurrency on each worker
  mygrep -p ".*panic.*" -f large.log -workers "node1:9090,node2:9090" -c 8

Note: Workers must be running before executing the master.
`)
}
