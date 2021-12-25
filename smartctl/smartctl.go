package smartctl

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/stretchr/objx"
)

const DefaultCommandTimeout = 15 * time.Second

type Command struct {
	smartctlBinary string
	smartctlArgs   []string
	useSudo        bool
	timeout        time.Duration
}

type CommandOption func(*Command)

func WithSudoEnabled() CommandOption {
	return func(c *Command) {
		c.useSudo = true
	}
}

func WithTimeout(t time.Duration) CommandOption {
	return func(c *Command) {
		c.timeout = t
	}
}

func WithSmartctlBinary(binaryPath string) CommandOption {
	return func(c *Command) {
		c.smartctlBinary = binaryPath
	}
}

func NewCommand(opts ...CommandOption) *Command {
	cmd := &Command{
		smartctlBinary: "smartctl",
		smartctlArgs:   []string{"--json", "-x"},
		useSudo:        false,
		timeout:        DefaultCommandTimeout,
	}

	for _, setOption := range opts {
		setOption(cmd)
	}

	return cmd
}

func (c *Command) QueryDevice(ctx context.Context, device string) (Data, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	binary := c.smartctlBinary
	baseArgs := c.smartctlArgs
	if c.useSudo {
		binary = "sudo"
		baseArgs = append([]string{c.smartctlBinary}, c.smartctlArgs...)
	}
	cmd := exec.CommandContext(ctx, binary, append(baseArgs, device)...)

	rawBytes, err := cmd.CombinedOutput()
	output := string(rawBytes)
	if err != nil {
		return Data{}, fmt.Errorf("command %s failed: %w", cmd.String(), richError(output, err))
	}

	raw, err := objx.FromJSON(output)
	if err != nil {
		return Data{}, err
	}
	return NewData(raw)
}

func richError(output string, err error) error {
	// No output, return process error
	if len(output) == 0 {
		return err
	}
	// fallback to dumping raw output
	fallbackErr := fmt.Errorf("%s (error: %w)", output, err)

	// Not a JSON output generated by Smartctl
	if output[0] != '{' {
		return fallbackErr
	}

	// Try to decode JSON
	raw, err := objx.FromJSON(output)
	if err != nil || !raw.Has("messages") {
		return fallbackErr
	}

	messages := []string{}
	raw.Get("messages").EachObjxMap(func(_ int, m objx.Map) bool {
		msg := m.Get("string").String()
		if msg != "" {
			messages = append(messages, msg)
		}
		return true
	})

	if len(messages) == 0 {
		return fallbackErr
	}
	return errors.New(strings.Join(messages, "; "))
}