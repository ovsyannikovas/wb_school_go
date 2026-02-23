package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Обработка сигналов
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		for range sigCh {
			fmt.Println("\nДля выхода используйте Ctrl+D")
		}
	}()

	for {
		fmt.Print("minishell> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("\nВыход из shell")
				return
			}
			fmt.Println("Ошибка ввода:", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Подстановка переменных окружения
		input = os.ExpandEnv(input)

		executeCommands(input)
	}
}

func executeCommands(input string) {
	// Разделяем по || и && для условного выполнения
	if strings.Contains(input, "||") {
		commands := strings.Split(input, "||")
		for _, cmd := range commands {
			if executeSingleCommand(strings.TrimSpace(cmd)) {
				// Если команда успешна, остальные не выполняем
				return
			}
		}
	} else if strings.Contains(input, "&&") {
		commands := strings.Split(input, "&&")
		for _, cmd := range commands {
			if !executeSingleCommand(strings.TrimSpace(cmd)) {
				// Если команда неуспешна, остальные не выполняем
				return
			}
		}
	} else {
		executeSingleCommand(input)
	}
}

func executeSingleCommand(input string) bool {
	var outputFile *os.File
	var inputFile *os.File

	// Проверка на > (вывод в файл)
	if strings.Contains(input, ">") {
		parts := strings.SplitN(input, ">", 2)
		input = strings.TrimSpace(parts[0])
		filename := strings.TrimSpace(parts[1])

		var err error
		outputFile, err = os.Create(filename)
		if err != nil {
			fmt.Printf("Ошибка создания файла: %v\n", err)
			return false
		}
		defer outputFile.Close()
	}

	// Проверка на < (ввод из файла)
	if strings.Contains(input, "<") {
		parts := strings.SplitN(input, "<", 2)
		input = strings.TrimSpace(parts[0])
		filename := strings.TrimSpace(parts[1])

		var err error
		inputFile, err = os.Open(filename)
		if err != nil {
			fmt.Printf("Ошибка открытия файла: %v\n", err)
			return false
		}
		defer inputFile.Close()
	}

	// Обработка конвейеров
	if strings.Contains(input, "|") {
		commands := strings.Split(input, "|")
		var cmdArgs [][]string

		for _, cmd := range commands {
			args := strings.Fields(strings.TrimSpace(cmd))
			if len(args) > 0 {
				cmdArgs = append(cmdArgs, args)
			}
		}

		if len(cmdArgs) > 0 {
			err := executePipeline(cmdArgs, inputFile, outputFile)
			return err == nil
		}
		return false
	}

	args := strings.Fields(input)
	if len(args) == 0 {
		return true
	}

	return executeSimpleCommand(args, inputFile, outputFile)
}

func executeSimpleCommand(args []string, inputFile, outputFile *os.File) bool {
	cmd := args[0]

	switch cmd {
	case "cd":
		if len(args) < 2 {
			fmt.Println("cd: требуется путь")
			return false
		}
		err := os.Chdir(args[1])
		if err != nil {
			fmt.Printf("cd: %v\n", err)
			return false
		}
		return true

	case "pwd":
		dir, err := os.Getwd()
		if err != nil {
			fmt.Printf("pwd: %v\n", err)
			return false
		}
		fmt.Println(dir)
		return true

	case "echo":
		fmt.Println(strings.Join(args[1:], " "))
		return true

	case "kill":
		if len(args) < 2 {
			fmt.Println("kill: требуется PID")
			return false
		}
		pid, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("kill: неверный PID %v\n", args[1])
			return false
		}
		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Printf("kill: процесс %d не найден\n", pid)
			return false
		}
		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			fmt.Printf("kill: %v\n", err)
			return false
		}
		return true

	case "ps":
		cmd := exec.Command("ps", "aux")
		cmd.Stdin = os.Stdin
		if outputFile != nil {
			cmd.Stdout = outputFile
		} else {
			cmd.Stdout = os.Stdout
		}
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		return err == nil

	default:
		return runExternalCommand(args, inputFile, outputFile)
	}
}

func runExternalCommand(args []string, inputFile, outputFile *os.File) bool {
	cmd := exec.Command(args[0], args[1:]...)

	if inputFile != nil {
		cmd.Stdin = inputFile
	} else {
		cmd.Stdin = os.Stdin
	}

	if outputFile != nil {
		cmd.Stdout = outputFile
	} else {
		cmd.Stdout = os.Stdout
	}

	cmd.Stderr = os.Stderr

	err := cmd.Run()
	return err == nil
}

func executePipeline(commands [][]string, inputFile, outputFile *os.File) error {
	if len(commands) == 0 {
		return nil
	}

	var cmds []*exec.Cmd

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmds = append(cmds, cmd)
	}

	for i := 0; i < len(cmds)-1; i++ {
		stdout, err := cmds[i].StdoutPipe()
		if err != nil {
			return err
		}
		// stdout текущей = stdin для след
		cmds[i+1].Stdin = stdout
	}

	// Настройка ввода для первой команды
	if inputFile != nil {
		cmds[0].Stdin = inputFile
	} else {
		cmds[0].Stdin = os.Stdin
	}

	// Настройка вывода для последней команды
	if outputFile != nil {
		cmds[len(cmds)-1].Stdout = outputFile
	} else {
		cmds[len(cmds)-1].Stdout = os.Stdout
	}

	for _, cmd := range cmds {
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return err
		}
	}

	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return err
		}
	}

	return nil
}
