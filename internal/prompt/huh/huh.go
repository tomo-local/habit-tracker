package huh

import (
	charmhuh "github.com/charmbracelet/huh"

	"habit-tracker/internal/prompt"
)

type prompter struct{}

func New() prompt.Prompter {
	return &prompter{}
}

// circleTheme swaps the default "> " arrow cursor and "✓/•" prefixes for
// round icons (●/◉/○).
func circleTheme() *charmhuh.Theme {
	t := charmhuh.ThemeCharm()
	t.Focused.SelectSelector = t.Focused.SelectSelector.SetString("● ")
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.SetString("● ")
	t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.SetString("◉ ")
	t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.SetString("○ ")
	t.Blurred.SelectedPrefix = t.Blurred.SelectedPrefix.SetString("◉ ")
	t.Blurred.UnselectedPrefix = t.Blurred.UnselectedPrefix.SetString("○ ")
	return t
}

func (p *prompter) Select(message string, options []string, defaultOption string) (string, error) {
	result := defaultOption
	if result == "" && len(options) > 0 {
		result = options[0]
	}
	err := charmhuh.NewForm(
		charmhuh.NewGroup(
			charmhuh.NewSelect[string]().
				Title(message).
				Description("/ to search").
				Options(charmhuh.NewOptions(options...)...).
				Value(&result),
		),
	).WithTheme(circleTheme()).Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (p *prompter) MultiSelect(message string, options []string) ([]string, error) {
	var result []string
	err := charmhuh.NewForm(
		charmhuh.NewGroup(
			charmhuh.NewMultiSelect[string]().
				Title(message).
				Options(charmhuh.NewOptions(options...)...).
				Value(&result),
		),
	).WithTheme(circleTheme()).Run()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *prompter) Input(message, defaultValue string) (string, error) {
	result := defaultValue
	err := charmhuh.NewForm(
		charmhuh.NewGroup(
			charmhuh.NewInput().
				Title(message).
				Value(&result),
		),
	).WithTheme(circleTheme()).Run()
	if err != nil {
		return "", err
	}
	if result == "" {
		result = defaultValue
	}
	return result, nil
}
