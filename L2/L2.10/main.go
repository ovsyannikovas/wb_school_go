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

type SortOptions struct {
	Column       int    // -k N (1-based, 0 означает всю строку)
	Numeric      bool   // -n
	Reverse      bool   // -r
	Unique       bool   // -u
	Month        bool   // -M
	IgnoreBlanks bool   // -b
	Check        bool   // -c
	HumanNumeric bool   // -h
	InputFile    string // имя файла или пустая строка для stdin
}

func sortLines(lines []string, opts *SortOptions) ([]string, error) {
	sort.Slice(lines, func(i, j int) bool {
		return compare(i, j, lines, opts)
	})

	if opts.Unique {
		var result []string
		for i, line := range lines {
			if i == 0 {
				result = append(result, line)
				continue
			}

			prev := lines[i-1]
			curr := line

			if opts.IgnoreBlanks {
				prev = trimTrailingSpaces(prev)
				curr = trimTrailingSpaces(curr)
			}

			if prev != curr {
				result = append(result, line)
			}
		}
		lines = result
	}

	return lines, nil
}

func compare(i, j int, lines []string, opts *SortOptions) bool {
	a := lines[i]
	b := lines[j]

	if opts.IgnoreBlanks {
		a = trimTrailingSpaces(a)
		b = trimTrailingSpaces(b)
	}
	if opts.Column > 0 {
		a = getColumn(a, opts.Column)
		b = getColumn(b, opts.Column)
	}

	if opts.Numeric {
		numA, errA := parseNumber(a, opts.IgnoreBlanks)
		numB, errB := parseNumber(b, opts.IgnoreBlanks)

		if errA == nil && errB == nil {
			if opts.Reverse {
				return numA > numB
			}
			return numA < numB
		}
		if errA != nil && errB == nil {
			return true
		}
		if errA == nil && errB != nil {
			return false
		}
	} else if opts.HumanNumeric {
		numA, errA := parseHumanNumber(a, opts.IgnoreBlanks)
		numB, errB := parseHumanNumber(b, opts.IgnoreBlanks)

		if errA == nil && errB == nil {
			if opts.Reverse {
				return numA > numB
			}
			return numA < numB
		}
		if errA != nil && errB == nil {
			return true
		}
		if errA == nil && errB != nil {
			return false
		}
	} else if opts.Month {
		monA, errA := parseMonth(a, opts.IgnoreBlanks)
		monB, errB := parseMonth(b, opts.IgnoreBlanks)

		if errA == nil && errB == nil {
			if opts.Reverse {
				return monA > monB
			}
			return monA < monB
		}
		if errA != nil && errB == nil {
			return true
		}
		if errA == nil && errB != nil {
			return false
		}
	}

	if opts.Reverse {
		return a > b
	}
	return a < b
}

func trimTrailingSpaces(s string) string {
	return strings.TrimRight(s, " \t")
}

func getColumn(line string, colNum int) string {
	cols := strings.Split(line, "\t")
	if colNum-1 < len(cols) {
		return cols[colNum-1]
	}
	return ""
}

func parseNumber(s string, ignoreBlanks bool) (float64, error) {
	if ignoreBlanks {
		s = strings.TrimRight(s, " \t")
	}

	return strconv.ParseFloat(s, 64)
}

func parseHumanNumber(s string, ignoreBlanks bool) (float64, error) {
	if ignoreBlanks {
		s = strings.TrimRight(s, " \t")
	}
	suffixPos := len(s) - 1
	for suffixPos >= 0 {
		c := s[suffixPos]
		if c >= '0' && c <= '9' || c == '.' || c == '-' || c == '+' {
			break
		}
		suffixPos--
	}

	if suffixPos < 0 {
		return strconv.ParseFloat(s, 64)
	}

	numberPart := s[:suffixPos+1]
	suffixPart := s[suffixPos+1:]

	num, err := strconv.ParseFloat(numberPart, 64)
	if err != nil {
		return 0, err
	}

	switch strings.ToUpper(suffixPart) {
	case "K":
		return num * 1000, nil
	case "M":
		return num * 1000 * 1000, nil
	case "G":
		return num * 1000 * 1000 * 1000, nil
	case "T":
		return num * 1000 * 1000 * 1000 * 1000, nil
	default:
		return 0, fmt.Errorf("unknown suffix: %s", suffixPart)
	}
}

func parseMonth(s string, ignoreBlanks bool) (int, error) {
	if ignoreBlanks {
		s = strings.TrimRight(s, " \t")
	}

	fields := strings.Fields(s)
	if len(fields) == 0 {
		return 0, fmt.Errorf("empty string")
	}

	for _, field := range fields {
		monthLower := strings.ToLower(field)

		months := map[string]int{
			"jan": 1, "feb": 2, "mar": 3, "apr": 4,
			"may": 5, "jun": 6, "jul": 7, "aug": 8,
			"sep": 9, "oct": 10, "nov": 11, "dec": 12,
		}

		if val, ok := months[monthLower]; ok {
			return val, nil
		}
	}

	return 0, fmt.Errorf("unknown month: %s", s)
}

func checkSorted(lines []string, opts *SortOptions) {
	for i := 1; i < len(lines); i++ {
		// Если предыдущая строка должна идти после текущей - порядок нарушен
		if compare(i, i-1, lines, opts) {
			fmt.Fprintf(os.Stderr, "sort: disorder: %s\n", lines[i])
			return
		}
	}
	fmt.Println("sort: input is sorted")
}

func validateFlags(opts *SortOptions) error {
	conflictCount := 0
	if opts.Numeric {
		conflictCount++
	}
	if opts.Month {
		conflictCount++
	}
	if opts.HumanNumeric {
		conflictCount++
	}

	if conflictCount > 1 {
		return fmt.Errorf("cannot combine -n, -M, and -h flags")
	}

	if opts.Check && (opts.Unique || opts.Reverse) {
		return fmt.Errorf("-c cannot be used with -u or -r")
	}

	return nil
}

func parseFlags() (*SortOptions, error) {
	sortOptions := &SortOptions{
		Column: 0,
	}

	fs := flag.NewFlagSet("sort", flag.ContinueOnError)

	kPtr := fs.Int("k", 0, "sort by column N")
	nPtr := fs.Bool("n", false, "numeric sort")
	rPtr := fs.Bool("r", false, "reverse order")
	uPtr := fs.Bool("u", false, "unique lines")
	mPtr := fs.Bool("M", false, "month sort")
	bPtr := fs.Bool("b", false, "ignore trailing blanks")
	cPtr := fs.Bool("c", false, "check if sorted")
	hPtr := fs.Bool("h", false, "human numeric sort")

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	if *kPtr > 0 {
		sortOptions.Column = *kPtr
	}
	sortOptions.Numeric = *nPtr
	sortOptions.Reverse = *rPtr
	sortOptions.Unique = *uPtr
	sortOptions.Month = *mPtr
	sortOptions.IgnoreBlanks = *bPtr
	sortOptions.Check = *cPtr
	sortOptions.HumanNumeric = *hPtr

	args := fs.Args()
	if len(args) > 0 {
		sortOptions.InputFile = args[0]
	}

	if err := validateFlags(sortOptions); err != nil {
		return nil, err
	}

	return sortOptions, nil
}

func readLines(opts *SortOptions) ([]string, error) {
	var lines []string
	var scanner *bufio.Scanner

	if opts.InputFile != "" {
		// file
		file, err := os.Open(opts.InputFile)
		if err != nil {
			return nil, fmt.Errorf("cannot open file %s: %v", opts.InputFile, err)
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	} else {
		// stdin
		scanner = bufio.NewScanner(os.Stdin)
	}

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
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

	lines, err := readLines(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	if opts.Check {
		checkSorted(lines, opts)
		return
	}

	sortedLines, err := sortLines(lines, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sorting: %v\n", err)
		os.Exit(1)
	}

	for _, line := range sortedLines {
		fmt.Println(line)
	}
}
