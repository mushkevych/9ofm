package controller

// Controller defines a renderable terminal screen pane.
type Renderer interface {
	Update() error
	Render() error
	IsVisible() bool
	SetVisible(bool) error
}

type Keymapper interface {
	// Keymap return string with all the keyboard actions a user can take
	// while the current pane is selected
	Keymap() string
}
