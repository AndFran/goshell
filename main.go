package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

type Command struct {
	Cmd  string
	Args []string
}

type History struct {
	Entry []string `json:"entry"`
}

func cleanString(s string) string {
	return strings.TrimSpace(s)
}

func tokenizeCommand(input string) (string, []string) {
	split := strings.Fields(input)
	command := cleanString(split[0])
	args := split[1:]
	for i, arg := range args {
		args[i] = cleanString(arg)
	}
	return command, args
}

func parseCommands(input string) []Command {
	commands := make([]Command, 0)
	pipeSplit := strings.Split(input, "|")
	for _, str := range pipeSplit {
		command, args := tokenizeCommand(str)
		commands = append(commands, Command{command, args})
	}

	return commands
}

func executeCommand(commands []Command) error {
	allCommands := make([]*exec.Cmd, 0)
	for i, command := range commands {
		cmd := exec.Command(command.Cmd, command.Args...)
		allCommands = append(allCommands, cmd)
		if i > 0 {
			out, err := allCommands[i-1].StdoutPipe()
			if err != nil {
				return err
			}
			cmd.Stdin = out
		}
	}

	allCommands[len(allCommands)-1].Stdout = os.Stdout // last command to console
	for _, command := range allCommands {
		err := command.Start()
		if err != nil {
			return err
		}
	}

	for _, command := range allCommands {
		err := command.Wait()
		if err != nil {
			return err
		}
	}
	return nil
}
func handleDirChange(arg string) error {
	var path string
	if len(arg) == 0 {
		path = "."
	} else {
		path = arg
	}
	err := os.Chdir(path)
	return err
}

func loadHistory() (*History, error) {
	historyFile, err := os.OpenFile("history.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	defer historyFile.Close()
	var history History
	err = json.NewDecoder(historyFile).Decode(&history)
	if err != nil {
		if err == io.EOF {
			return &History{Entry: make([]string, 0)}, nil
		}
		return nil, err
	}

	return &history, nil
}

func saveHistory(history *History) error {
	file, err := os.OpenFile("history.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(history)
	if err != nil {
		return err
	}
	return nil
}

func showHistory(history *History) {
	for _, h := range history.Entry {
		fmt.Println(h)
	}
}

func main() {
	signal.Ignore(os.Interrupt)
	var err error

	history, err := loadHistory()
	if err != nil {
		log.Fatal(err)
	}

	for {
		fmt.Print("ccsh>")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		commands := parseCommands(input)

		if len(commands) == 0 {
			fmt.Println("No commands found.")
			continue
		}

		switch commands[0].Cmd {

		case "cd":
			if len(commands[0].Args) == 0 {
				homeDir, _ := os.UserHomeDir()
				commands[0].Args = append(commands[0].Args, homeDir)
			}
			err = handleDirChange(commands[0].Args[0])
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "exit":
			err = saveHistory(history)
			if err != nil {
				fmt.Println("Error saving history:", err)
			}
			os.Exit(0)
		case "pwd":
			pwd, _ := os.Getwd()
			fmt.Println(pwd)
		case "history":
			showHistory(history)
		default:
			err = executeCommand(commands)
			if err != nil {
				fmt.Println("Unknown command")
			}
		}
		history.Entry = append(history.Entry, strings.TrimSpace(input))
	}
}
