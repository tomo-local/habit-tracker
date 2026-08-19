package prompt

// Option is a selectable item for Select. Label is what's shown to the
// user; Value is what's returned when this option is chosen, and need not
// equal Label (e.g. a display name paired with a stable ID).
type Option struct {
	Label string
	Value string
}

type Prompter interface {
	// Select prompts the user to choose one of options and returns the
	// chosen Option. defaultValue is matched against Option.Value to
	// determine the initial selection.
	Select(message string, options []Option, defaultValue string) (Option, error)
	MultiSelect(message string, options []string) ([]string, error)
	Input(message, defaultValue string) (string, error)
}
