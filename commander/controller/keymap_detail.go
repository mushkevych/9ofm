package controller

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/mushkevych/9ofm/commander/format"
	"github.com/wagoodman/keybinding"
)

type KeymapDetail struct {
	// gocui.Key are not unique: for instance KeyEnter and KeyCtrlM share the same value 0x0D
	// hence we need to provide and keep "human readable keyboard shortcut" such as "ctrl+m"
	KeyboardShortcut string

	// Display represents name of the action shown to the user. For instance: Copy, Quit, Edit
	Display string

	OnAction   func() error
	IsSelected func() bool
}

func RegisterKeymaps(gui *gocui.Gui, viewname string, keymaps []KeymapDetail) error {
	for _, keymap := range keymaps {
		key, err := keybinding.Parse(keymap.KeyboardShortcut)
		if err != nil {
			return err
		}

		// a keymap deep copy is required, as otherwise `keymap.onAction` pointer
		// is recycled and points to the last element of the loop
		var kmCopy = keymap.deepCopy()
		if err := gui.SetKeybinding(viewname, key.Value, key.Modifier, kmCopy.onAction); err != nil {
			return err
		}
	}
	return nil
}

func (km *KeymapDetail) deepCopy() KeymapDetail {
	return KeymapDetail{
		KeyboardShortcut: km.KeyboardShortcut,
		Display:          km.Display,
		OnAction:         km.OnAction,
		IsSelected:       km.IsSelected,
	}
}

func (km *KeymapDetail) onAction(*gocui.Gui, *gocui.View) error {
	if km.OnAction == nil {
		return fmt.Errorf("no action configured for '%+v'", km)
	}
	return km.OnAction()
}

// isVisible returns *true* if the UI component associated with this Keymap is visible
func (km *KeymapDetail) isVisible() bool {
	if km.IsSelected == nil {
		return false
	}
	return km.IsSelected()
}

func (km *KeymapDetail) String() string {
	// abbreviation is a "human readable shortcut" such as "^M" for "ctrl+m"
	var key keybinding.Key
	key, _ = keybinding.Parse(km.KeyboardShortcut)
	abbreviation := key.String()

	return format.FmtKeymap(abbreviation, km.Display, km.isVisible())
}
