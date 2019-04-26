package terminal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

// GetInput returns some user entered text or an error
func GetInput(prompt string) (string, error) {

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("/n" + prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", errors.Wrap(err, "error reading input")
	}

	return strings.Replace(input, "\n", "", -1), nil
}

// GetInputSecret returns some user entered text or an error.
// The user text is not echoed to the terminal.
func GetInputSecret(prompt string) (string, error) {

	fmt.Print("/n" + prompt)
	input, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", errors.Wrap(err, "error reading input")
	}

	return string(input), nil
}
