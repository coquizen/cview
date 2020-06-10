package cview

import (
	"sync"

	"github.com/gdamore/tcell"
)

// Modal is a centered message window used to inform the user or prompt them
// for an immediate decision. It needs to have at least one button (added via
// AddButtons) or it will never disappear. You may include additional elements
// within the window by modifying the Form returned by GetForm.
type Modal struct {
	*Box

	// The Frame embedded in the Modal.
	frame *Frame

	// The Form embedded in the Modal's Frame.
	form *Form

	// The message text (original, not word-wrapped).
	text string

	// The text color.
	textColor tcell.Color

	// The optional callback for when the user clicked one of the buttons. It
	// receives the index of the clicked button and the button's label.
	done func(buttonIndex int, buttonLabel string)

	sync.Mutex
}

// NewModal returns a new centered message window.
func NewModal() *Modal {
	m := &Modal{
		Box:       NewBox(),
		textColor: Styles.PrimaryTextColor,
	}
	m.form = NewForm().
		SetButtonsAlign(AlignCenter).
		SetButtonBackgroundColor(Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(Styles.PrimaryTextColor)
	m.form.SetBackgroundColor(Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
	m.form.SetCancelFunc(func() {
		if m.done != nil {
			m.done(-1, "")
		}
	})
	m.frame = NewFrame(m.form).SetBorders(0, 0, 1, 0, 0, 0)
	m.frame.SetBorder(true).
		SetBackgroundColor(Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)
	m.focus = m
	return m
}

// SetBackgroundColor sets the color of the Modal Frame background.
func (m *Modal) SetBackgroundColor(color tcell.Color) *Modal {
	m.Lock()
	defer m.Unlock()

	m.form.SetBackgroundColor(color)
	m.frame.SetBackgroundColor(color)
	return m
}

// SetTextColor sets the color of the message text.
func (m *Modal) SetTextColor(color tcell.Color) *Modal {
	m.Lock()
	defer m.Unlock()

	m.textColor = color
	return m
}

// SetButtonBackgroundColor sets the background color of the buttons.
func (m *Modal) SetButtonBackgroundColor(color tcell.Color) *Modal {
	m.Lock()
	defer m.Unlock()

	m.form.SetButtonBackgroundColor(color)
	return m
}

// SetButtonTextColor sets the color of the button texts.
func (m *Modal) SetButtonTextColor(color tcell.Color) *Modal {
	m.Lock()
	defer m.Unlock()

	m.form.SetButtonTextColor(color)
	return m
}

// SetDoneFunc sets a handler which is called when one of the buttons was
// pressed. It receives the index of the button as well as its label text. The
// handler is also called when the user presses the Escape key. The index will
// then be negative and the label text an empty string.
func (m *Modal) SetDoneFunc(handler func(buttonIndex int, buttonLabel string)) *Modal {
	m.Lock()
	defer m.Unlock()

	m.done = handler
	return m
}

// SetText sets the message text of the window. The text may contain line
// breaks. Note that words are wrapped, too, based on the final size of the
// window.
func (m *Modal) SetText(text string) *Modal {
	m.Lock()
	defer m.Unlock()

	m.text = text
	return m
}

// GetForm returns the Form embedded in the window. The returned Form may be
// modified to include additional elements (e.g. AddInputField, AddFormItem).
func (m *Modal) GetForm() *Form {
	m.Lock()
	defer m.Unlock()

	return m.form
}

// GetFrame returns the Frame embedded in the window.
func (m *Modal) GetFrame() *Frame {
	m.Lock()
	defer m.Unlock()

	return m.frame
}

// AddButtons adds buttons to the window. There must be at least one button and
// a "done" handler so the window can be closed again.
func (m *Modal) AddButtons(labels []string) *Modal {
	m.Lock()
	defer m.Unlock()

	for index, label := range labels {
		func(i int, l string) {
			m.form.AddButton(label, func() {
				if m.done != nil {
					m.done(i, l)
				}
			})
			button := m.form.GetButton(m.form.GetButtonCount() - 1)
			button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyDown, tcell.KeyRight:
					return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
				case tcell.KeyUp, tcell.KeyLeft:
					return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
				}
				return event
			})
		}(index, label)
	}
	return m
}

// ClearButtons removes all buttons from the window.
func (m *Modal) ClearButtons() *Modal {
	m.Lock()
	defer m.Unlock()

	m.form.ClearButtons()
	return m
}

// SetFocus shifts the focus to the button with the given index.
func (m *Modal) SetFocus(index int) *Modal {
	m.Lock()
	defer m.Unlock()

	m.form.SetFocus(index)
	return m
}

// Focus is called when this primitive receives focus.
func (m *Modal) Focus(delegate func(p Primitive)) {
	delegate(m.form)
}

// HasFocus returns whether or not this primitive has focus.
func (m *Modal) HasFocus() bool {
	m.Lock()
	defer m.Unlock()

	return m.form.HasFocus()
}

// Draw draws this primitive onto the screen.
func (m *Modal) Draw(screen tcell.Screen) {
	formItemCount := m.form.GetFormItemCount()

	m.Lock()
	defer m.Unlock()

	// Calculate the width of this Modal.
	buttonsWidth := 0
	for _, button := range m.form.buttons {
		buttonsWidth += TaggedStringWidth(button.label) + 4 + 2
	}
	buttonsWidth -= 2
	screenWidth, screenHeight := screen.Size()
	width := screenWidth / 3
	if width < buttonsWidth {
		width = buttonsWidth
	}
	// width is now without the box border.

	// Reset the text and find out how wide it is.
	m.frame.Clear()
	lines := WordWrap(m.text, width)
	for _, line := range lines {
		m.frame.AddText(line, true, AlignCenter, m.textColor)
	}

	// Set the Modal's position and size.
	height := len(lines) + (formItemCount * 2) + 6
	width += 4
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2
	m.SetRect(x, y, width, height)

	// Draw the frame.
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}

// MouseHandler returns the mouse handler for this primitive.
func (m *Modal) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return m.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		// Pass mouse events on to the form.
		consumed, capture = m.form.MouseHandler()(action, event, setFocus)
		if !consumed && action == MouseLeftClick && m.InRect(event.Position()) {
			setFocus(m)
			consumed = true
		}
		return
	})
}
