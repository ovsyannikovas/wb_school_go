package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type CutOptions struct {
	Fields    string // -f (номера полей (строк), которые нужно вывести)
	Delimiter string // -d (другой разделитель)
	Separated bool   // -s (только строки, содержащие разделитель)
}

func cut(line string, opts *CutOptions, fields []int) (string, error) {
	if opts.Delimiter == "" {
		return "", fmt.Errorf("delimiter cannot be empty")
	}

	// Проверяем наличие разделителя для флага -s
	if opts.Separated && !strings.Contains(line, opts.Delimiter) {
		return "", nil
	}

	parts := strings.Split(line, opts.Delimiter)

	if len(fields) == 0 {
		return line, nil
	}

	var selected []string
	for _, fieldNum := range fields {
		idx := fieldNum - 1
		if idx >= 0 && idx < len(parts) {
			selected = append(selected, parts[idx])
		}
	}

	return strings.Join(selected, opts.Delimiter), nil
}

func parseFlags() (*CutOptions, error) {
	opts := &CutOptions{}

	fieldsPtr := flag.String("f", "", "номера полей для вывода (например: 1,3-5)")
	delimiterPtr := flag.String("d", "\t", "разделитель полей (по умолчанию табуляция)")
	separatedPtr := flag.Bool("s", false, "выводить только строки с разделителем")

	flag.Parse()

	opts.Fields = *fieldsPtr
	opts.Delimiter = *delimiterPtr
	opts.Separated = *separatedPtr

	if opts.Fields == "" {
		return nil, fmt.Errorf("fields specification (-f) is required")
	}

	return opts, nil
}

func readLines() ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %v", err)
	}

	return lines, nil
}

func main() {
	opts, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var fields []int
	if opts.Fields != "" {
		fields, err = parseFields(opts.Fields)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing fields: %v\n", err)
			os.Exit(1)
		}
	}

	lines, err := readLines()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	for _, line := range lines {
		result, err := cut(line, opts, fields)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing line: %v\n", err)
			continue
		}
		if result != "" {
			fmt.Println(result)
		}
	}
}

func parseFields(fields string) ([]int, error) {
	if fields == "" {
		return nil, fmt.Errorf("empty fields specification")
	}

	var result []int
	seen := make(map[int]bool)

	parts := strings.Split(fields, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			startStr := strings.TrimSpace(rangeParts[0])
			endStr := strings.TrimSpace(rangeParts[1])

			if startStr == "" || endStr == "" {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err := strconv.Atoi(startStr)
			if err != nil {
				return nil, fmt.Errorf("invalid number in range: %s", startStr)
			}

			end, err := strconv.Atoi(endStr)
			if err != nil {
				return nil, fmt.Errorf("invalid number in range: %s", endStr)
			}

			if start < 1 || end < 1 {
				return nil, fmt.Errorf("field numbers must be positive: %d-%d", start, end)
			}

			if start > end {
				return nil, fmt.Errorf("invalid range: start > end (%d > %d)", start, end)
			}

			for i := start; i <= end; i++ {
				if !seen[i] {
					result = append(result, i)
					seen[i] = true
				}
			}
		} else {
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}

			if num < 1 {
				return nil, fmt.Errorf("field number must be positive: %d", num)
			}

			if !seen[num] {
				result = append(result, num)
				seen[num] = true
			}
		}
	}

	sort.Ints(result)

	return result, nil
}
