package terminal

import (
	"github.com/manifoldco/promptui"
)

// TextGatherer displays a prompt and harvests textual input
type TextGatherer func(label string) (string, error)

// GetInput returns a function to return some user entered text or an error
func GetInput() TextGatherer {
	return func(label string) (string, error) {

		prompt := promptui.Prompt{
			Label: label,
		}
		return prompt.Run()
	}
}

// GetInputSecret returns a function to return some user entered text or an error.
// The user text is not echoed to the terminal.
func GetInputSecret() TextGatherer {
	return func(label string) (string, error) {

		prompt := promptui.Prompt{
			Label: label,
			Mask:  '*',
		}
		return prompt.Run()
	}
}
