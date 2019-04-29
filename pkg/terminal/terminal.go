package terminal

import (
	"github.com/manifoldco/promptui"
)

// GetInput returns some user entered text or an error
func GetInput(label string) (string, error) {

	prompt := promptui.Prompt{
		Label: label,
	}

	return prompt.Run()
}

// GetInputSecret returns some user entered text or an error.
// The user text is not echoed to the terminal.
func GetInputSecret(label string) (string, error) {

	prompt := promptui.Prompt{
		Label: label,
		Mask:  '*',
	}

	return prompt.Run()
}
