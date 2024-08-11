package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Command struct {
	Cmd  string
	Args []string
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

func main() {
	var err error
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
			os.Exit(0)
		case "pwd":
			pwd, _ := os.Getwd()
			fmt.Println(pwd)
		default:
			err = executeCommand(commands)
			if err != nil {
				fmt.Println("Unknown command")
			}
		}
	}
}
