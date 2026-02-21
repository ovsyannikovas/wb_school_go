package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type GrepOptions struct {
	After      int    // -A N (строк после)
	Before     int    // -B N (строк до)
	Context    int    // -C N (контекст вокруг)
	Count      bool   // -c (только количество)
	IgnoreCase bool   // -i (игнорировать регистр)
	Invert     bool   // -v (инвертировать)
	Fixed      bool   // -F (точное совпадение, не regexp)
	LineNum    bool   // -n (показывать номера строк)
	Pattern    string // сам шаблон
	InputFile  string // имя файла или пустая строка для stdin
}

func match(line string, opts *GrepOptions, regex *regexp.Regexp) bool {
	if opts.Fixed {
		lineToCheck := line
		patternToCheck := opts.Pattern

		if opts.IgnoreCase {
			lineToCheck = strings.ToLower(line)
			patternToCheck = strings.ToLower(opts.Pattern)
		}

		contains := strings.Contains(lineToCheck, patternToCheck)

		if opts.Invert {
			return !contains
		}
		return contains
	}

	matched := regex.MatchString(line)

	if opts.Invert {
		return !matched
	}
	return matched
}

func parseFlags() (*GrepOptions, *regexp.Regexp, error) {
	opts := &GrepOptions{}

	afterPtr := flag.Int("A", 0, "print N lines after match")
	beforePtr := flag.Int("B", 0, "print N lines before match")
	contextPtr := flag.Int("C", 0, "print N lines around match")
	countPtr := flag.Bool("c", false, "print only count of matching lines")
	ignoreCasePtr := flag.Bool("i", false, "ignore case")
	invertPtr := flag.Bool("v", false, "invert match")
	fixedPtr := flag.Bool("F", false, "fixed string match (not regex)")
	lineNumPtr := flag.Bool("n", false, "print line numbers")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] PATTERN [FILE]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	opts.After = *afterPtr
	opts.Before = *beforePtr
	opts.Context = *contextPtr
	opts.Count = *countPtr
	opts.IgnoreCase = *ignoreCasePtr
	opts.Invert = *invertPtr
	opts.Fixed = *fixedPtr
	opts.LineNum = *lineNumPtr

	if opts.Context > 0 {
		opts.After = opts.Context
		opts.Before = opts.Context
	}

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return nil, nil, fmt.Errorf("missing pattern")
	}

	opts.Pattern = args[0]

	if len(args) > 1 {
		opts.InputFile = args[1]
	}

	if opts.After < 0 || opts.Before < 0 || opts.Context < 0 {
		return nil, nil, fmt.Errorf("context arguments must be non-negative")
	}

	if opts.Pattern == "" {
		return nil, nil, fmt.Errorf("pattern cannot be empty")
	}

	// Компилируем регулярное выражение, если нужно
	var regex *regexp.Regexp
	var err error
	if !opts.Fixed {
		pattern := opts.Pattern
		if opts.IgnoreCase {
			pattern = "(?i)" + pattern
		}
		regex, err = regexp.Compile(pattern)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid regular expression: %v", err)
		}
	}

	return opts, regex, nil
}

func readLines(filename string) ([]string, error) {
	var lines []string
	var scanner *bufio.Scanner

	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("cannot open file %s: %v", filename, err)
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	} else {
		scanner = bufio.NewScanner(os.Stdin)
	}

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %v", err)
	}

	return lines, nil
}

func main() {
	opts, regex, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	lines, err := readLines(opts.InputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Если нужен только подсчет
	if opts.Count {
		count := 0
		for _, line := range lines {
			if match(line, opts, regex) {
				count++
			}
		}
		fmt.Println(count)
		return
	}

	// Обработка инвертированного поиска (без контекста)
	if opts.Invert {
		for i, line := range lines {
			if match(line, opts, regex) {
				if opts.LineNum {
					fmt.Printf("%d:%s\n", i+1, line)
				} else {
					fmt.Println(line)
				}
			}
		}
		return
	}

	// Для контекста отмечаем, какие строки выводить
	outputLines := make(map[int]bool)

	// Находим все совпадения и добавляем контекст
	for i, line := range lines {
		if match(line, opts, regex) {
			start := max(0, i-opts.Before)
			end := min(len(lines)-1, i+opts.After)
			for j := start; j <= end; j++ {
				outputLines[j] = true
			}
		}
	}

	// Выводим результат с разделителями
	var lastOutput int = -2
	for i := 0; i < len(lines); i++ {
		if outputLines[i] {
			if lastOutput != -2 && i > lastOutput+1 {
				fmt.Println("--")
			}

			if opts.LineNum {
				fmt.Printf("%d:%s\n", i+1, lines[i])
			} else {
				fmt.Println(lines[i])
			}
			lastOutput = i
		}
	}
}
