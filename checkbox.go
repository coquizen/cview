package cview

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

// CheckBox implements a simple box for boolean values which can be checked and
// unchecked.
type CheckBox struct {
	*Box

	// Whether or not this box is checked.
	checked bool

	// The text to be displayed before the checkbox.
	label string

	// The text to be displayed after the checkbox.
	message string

	// The screen width of the label area. A value of 0 means use the width of
	// the label text.
	labelWidth int

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// An optional function which is called when the user changes the checked
	// state of this checkbox.
	changed func(checked bool)

	// An optional function which is called when the user indicated that they
	// are done entering text. The key which was pressed is provided (tab,
	// shift-tab, or escape).
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)

	// The rune to show when the checkbox is checked
	checkedRune rune

	sync.RWMutex
}

// NewCheckBox returns a new input field.
func NewCheckBox() *CheckBox {
	return &CheckBox{
		Box:                  NewBox(),
		labelColor:           Styles.SecondaryTextColor,
		fieldBackgroundColor: Styles.ContrastBackgroundColor,
		fieldTextColor:       Styles.PrimaryTextColor,
		checkedRune:          Styles.CheckBoxCheckedRune,
	}
}

// SetChecked sets the state of the checkbox.
func (c *CheckBox) SetChecked(checked bool) *CheckBox {
	c.Lock()
	defer c.Unlock()

	c.checked = checked
	return c
}

// SetCheckedRune sets the rune to show when the checkbox is checked.
func (c *CheckBox) SetCheckedRune(rune rune) *CheckBox {
	c.Lock()
	defer c.Unlock()

	c.checkedRune = rune
	return c
}

// IsChecked returns whether or not the box is checked.
func (c *CheckBox) IsChecked() bool {
	c.RLock()
	defer c.RUnlock()

	return c.checked
}

// SetLabel sets the text to be displayed before the input area.
func (c *CheckBox) SetLabel(label string) *CheckBox {
	c.Lock()
	defer c.Unlock()

	c.label = label
	return c
}

// GetLabel returns the text to be displayed before the input area.
func (c *CheckBox) GetLabel() string {
	c.RLock()
	defer c.RUnlock()

	return c.label
}

// SetMessage sets the text to be displayed after the checkbox
func (c *CheckBox) SetMessage(message string) *CheckBox {
	c.Lock()
	defer c.Unlock()

	c.message = message
	return c
}

// GetMessage returns the text to be displayed after the checkbox
func (c *CheckBox) GetMessage() string {
	c.RLock()
	defer c.RUnlock()

	return c.message
}

// SetLabelWidth sets the screen width of the label. A value of 0 will cause the
// primitive to use the width of the label string.
func (c *CheckBox) SetLabelWidth(width int) *CheckBox {
	c.Lock()
	defer c.Unlock()

	c.labelWidth = width
	return c
}

// SetLabelColor sets the color of the label.
func (c *CheckBox) SetLabelColor(color tcell.Color) *CheckBox {
	c.Lock()
	defer c.Unlock()

	c.labelColor = color
	return c
}

// SetFieldBackgroundColor sets the background color of the input area.
func (c *CheckBox) SetFieldBackgroundColor(color tcell.Color) *CheckBox {
	c.Lock()
	defer c.Unlock()

	c.fieldBackgroundColor = color
	return c
}

// SetFieldTextColor sets the text color of the input area.
func (c *CheckBox) SetFieldTextColor(color tcell.Color) *CheckBox {
	c.Lock()
	defer c.Unlock()

	c.fieldTextColor = color
	return c
}

// SetFormAttributes sets attributes shared by all form items.
func (c *CheckBox) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	c.Lock()
	defer c.Unlock()

	c.labelWidth = labelWidth
	c.labelColor = labelColor
	c.backgroundColor = bgColor
	c.fieldTextColor = fieldTextColor
	c.fieldBackgroundColor = fieldBgColor
	return c
}

// GetFieldWidth returns this primitive's field width.
func (c *CheckBox) GetFieldWidth() int {
	c.RLock()
	defer c.RUnlock()

	if c.message == "" {
		return 1
	}

	return 2 + len(c.message)
}

// SetChangedFunc sets a handler which is called when the checked state of this
// checkbox was changed by the user. The handler function receives the new
// state.
func (c *CheckBox) SetChangedFunc(handler func(checked bool)) *CheckBox {
	c.Lock()
	defer c.Unlock()

	c.changed = handler
	return c
}

// SetDoneFunc sets a handler which is called when the user is done using the
// checkbox. The callback function is provided with the key that was pressed,
// which is one of the following:
//
//   - KeyEscape: Abort text input.
//   - KeyTab: Move to the next field.
//   - KeyBacktab: Move to the previous field.
func (c *CheckBox) SetDoneFunc(handler func(key tcell.Key)) *CheckBox {
	c.Lock()
	defer c.Unlock()

	c.done = handler
	return c
}

// SetFinishedFunc sets a callback invoked when the user leaves this form item.
func (c *CheckBox) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	c.Lock()
	defer c.Unlock()

	c.finished = handler
	return c
}

// Draw draws this primitive onto the screen.
func (c *CheckBox) Draw(screen tcell.Screen) {
	c.Box.Draw(screen)

	c.Lock()
	defer c.Unlock()

	// Prepare
	x, y, width, height := c.GetInnerRect()
	rightLimit := x + width
	if height < 1 || rightLimit <= x {
		return
	}

	// Draw label.
	if c.labelWidth > 0 {
		labelWidth := c.labelWidth
		if labelWidth > rightLimit-x {
			labelWidth = rightLimit - x
		}
		Print(screen, c.label, x, y, labelWidth, AlignLeft, c.labelColor)
		x += labelWidth
	} else {
		_, drawnWidth := Print(screen, c.label, x, y, rightLimit-x, AlignLeft, c.labelColor)
		x += drawnWidth
	}

	// Draw checkbox.
	fieldStyle := tcell.StyleDefault.Background(c.fieldBackgroundColor).Foreground(c.fieldTextColor)
	if c.focus.HasFocus() {
		fieldStyle = fieldStyle.Background(c.fieldTextColor).Foreground(c.fieldBackgroundColor)
	}

	checkedRune := c.checkedRune
	if !c.checked {
		checkedRune = ' '
	}
	screen.SetContent(x, y, checkedRune, nil, fieldStyle)

	if c.message != "" {
		Print(screen, c.message, x+2, y, len(c.message), AlignLeft, c.labelColor)
	}
}

// InputHandler returns the handler for this primitive.
func (c *CheckBox) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return c.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if HitShortcut(event, Keys.Select, Keys.Select2) {
			c.Lock()
			c.checked = !c.checked
			c.Unlock()
			if c.changed != nil {
				c.changed(c.checked)
			}
		} else if HitShortcut(event, Keys.Cancel, Keys.MovePreviousField, Keys.MoveNextField) {
			if c.done != nil {
				c.done(event.Key())
			}
			if c.finished != nil {
				c.finished(event.Key())
			}
		}
	})
}

// MouseHandler returns the mouse handler for this primitive.
func (c *CheckBox) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return c.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		x, y := event.Position()
		_, rectY, _, _ := c.GetInnerRect()
		if !c.InRect(x, y) {
			return false, nil
		}

		// Process mouse event.
		if action == MouseLeftClick && y == rectY {
			setFocus(c)
			c.checked = !c.checked
			if c.changed != nil {
				c.changed(c.checked)
			}
			consumed = true
		}

		return
	})
}
