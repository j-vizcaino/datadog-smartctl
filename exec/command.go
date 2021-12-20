package exec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type ErrorHandler func(cmdLine string, exitCode int, output string)

// NoopHandler ignores errors
func NoopHandler(_ string, _ int, _ string) {}

func ExitOnErrorHandler(cmdLine string, exitCode int, output string) {
	fmt.Println(output)
	fmt.Printf("ERROR: %s command failed with exit code %d\n", cmdLine, exitCode)
	os.Exit(1)
}

var DefaultErrorHandler = ExitOnErrorHandler

type Command struct {
	command      *exec.Cmd
	errorHandler ErrorHandler
}

type CommandOption func(cmd *exec.Cmd)

func NewCommand(prog string, args ...string) *Command {
	return &Command{
		command:      exec.Command(prog, args...),
		errorHandler: DefaultErrorHandler,
	}
}

func (c *Command) WithErrorHandler(errHandler ErrorHandler) *Command {
	c.errorHandler = errHandler
	return c
}

func (c *Command) WithOption(opt CommandOption) *Command {
	opt(c.command)
	return c
}

func (c *Command) Run() string {
	var exitCode int
	var outbuf, errbuf bytes.Buffer
	if c.command.Stdout == nil {
		c.command.Stdout = &outbuf
	}
	if c.command.Stderr == nil {
		c.command.Stderr = &errbuf
	}

	err := c.command.Run()
	stdout := strings.TrimSuffix(outbuf.String(), "\n")
	stderr := strings.TrimSuffix(errbuf.String(), "\n")

	cmdLine := strings.Join(c.command.Args, " ")

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			exitCode = 127
		}
		c.errorHandler(cmdLine, exitCode, stderr)
	} else {
		// success, exitCode should be 0 if go is ok
		ws := c.command.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
		if exitCode != 0 {
			c.errorHandler(cmdLine, exitCode, stderr)
		}
	}
	return stdout
}
