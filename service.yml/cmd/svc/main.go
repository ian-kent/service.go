package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ian-kent/service.go/service.yml/def"
)

func main() {
	if len(os.Args) <= 1 {
		printUsage()
	}

	svc := def.MustLoadOne(def.DefaultNames...)

	command := os.Args[1]
	var args []string
	args = append(args, os.Args[:1]...)
	args = append(args, os.Args[2:]...)
	os.Args = args

	switch command {
	case "info":
		fmt.Printf("%+v\n", svc)
	default:
		if target, ok := svc.Targets[command]; ok {
			runCommand(target)
			return
		}
		printUsageWithError(fmt.Errorf("unknown command or target '%s'", command))
	}
}

func runCommand(command string) {
	var cmd *exec.Cmd
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case sig := <-sigs:
				cmd.Process.Signal(sig)
			}
		}
	}()

	// FIXME parse properly, quotes etc
	args := strings.Split(command, " ")

	cmd = exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			if status, ok := err.(*exec.ExitError).Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
			os.Exit(1)
		default:
			printError(fmt.Errorf("executing command: %s", err))
		}
	}
}

func printUsage() {
	printUsageWithError(nil)
}

func printUsageWithError(err error) {
	fmt.Fprintf(os.Stderr, "usage: svc <command> [parameters]\n")
	printError(err)
	os.Exit(0)
}

func printError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "svc: error: %s\n", err.Error())
		os.Exit(1)
	}
}
