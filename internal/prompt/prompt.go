package prompt

type Prompter interface {
	Select(message string, options []string, defaultOption string) (string, error)
	MultiSelect(message string, options []string) ([]string, error)
	Input(message, defaultValue string) (string, error)
}
