package cview

import (
	"math"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

// InputField is a one-line box (three lines if there is a title) where the
// user can enter text. Use SetAcceptanceFunc() to accept or reject input,
// SetChangedFunc() to listen for changes, and SetMaskCharacter() to hide input
// from onlookers (e.g. for password input).
//
// The following keys can be used for navigation and editing:
//
//   - Left arrow: Move left by one character.
//   - Right arrow: Move right by one character.
//   - Home, Ctrl-A, Alt-a: Move to the beginning of the line.
//   - End, Ctrl-E, Alt-e: Move to the end of the line.
//   - Alt-left, Alt-b: Move left by one word.
//   - Alt-right, Alt-f: Move right by one word.
//   - Backspace: Delete the character before the cursor.
//   - Delete: Delete the character after the cursor.
//   - Ctrl-K: Delete from the cursor to the end of the line.
//   - Ctrl-W: Delete the last word before the cursor.
//   - Ctrl-U: Delete the entire line.
type InputField struct {
	*Box

	// The text that was entered.
	text string

	// The text to be displayed before the input area.
	label string

	// The text to be displayed in the input area when "text" is empty.
	placeholder string

	// The label color.
	labelColor tcell.Color

	// The label color when focused.
	labelColorFocused tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The background color of the input area when focused.
	fieldBackgroundColorFocused tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// The text color of the input area when focused.
	fieldTextColorFocused tcell.Color

	// The text color of the placeholder.
	placeholderTextColor tcell.Color

	// The text color of the placeholder when focused.
	placeholderTextColorFocused tcell.Color

	// The text color of the list items.
	autocompleteListTextColor tcell.Color

	// The background color of the autocomplete list.
	autocompleteListBackgroundColor tcell.Color

	// The text color of the selected ListItem.
	autocompleteListSelectedTextColor tcell.Color

	// The background color of the selected ListItem.
	autocompleteListSelectedBackgroundColor tcell.Color

	// The text color of the suggestion.
	autocompleteSuggestionTextColor tcell.Color

	// The text color of the note below the input field.
	fieldNoteTextColor tcell.Color

	// The note to show below the input field.
	fieldNote string

	// The screen width of the label area. A value of 0 means use the width of
	// the label text.
	labelWidth int

	// The screen width of the input area. A value of 0 means extend as much as
	// possible.
	fieldWidth int

	// A character to mask entered text (useful for password fields). A value of 0
	// disables masking.
	maskCharacter rune

	// The cursor position as a byte index into the text string.
	cursorPos int

	// An optional autocomplete function which receives the current text of the
	// input field and returns a slice of ListItems to be displayed in a drop-down
	// selection. Items' main text is displayed in the autocomplete list. When
	// set, items' secondary text is used as the selection value. Otherwise,
	// the main text is used.
	autocomplete func(text string) []*ListItem

	// The List object which shows the selectable autocomplete entries. If not
	// nil, the list's main texts represent the current autocomplete entries.
	autocompleteList *List

	// The suggested completion of the current autocomplete ListItem.
	autocompleteListSuggestion string

	// An optional function which may reject the last character that was entered.
	accept func(text string, ch rune) bool

	// An optional function which is called when the input has changed.
	changed func(text string)

	// An optional function which is called when the user indicated that they
	// are done entering text. The key which was pressed is provided (tab,
	// shift-tab, enter, or escape).
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)

	// The x-coordinate of the input field as determined during the last call to Draw().
	fieldX int

	// The number of bytes of the text string skipped ahead while drawing.
	offset int

	sync.RWMutex
}

// NewInputField returns a new input field.
func NewInputField() *InputField {
	return &InputField{
		Box:                                     NewBox(),
		labelColor:                              Styles.SecondaryTextColor,
		fieldBackgroundColor:                    Styles.ContrastBackgroundColor,
		fieldTextColor:                          Styles.PrimaryTextColor,
		placeholderTextColor:                    Styles.ContrastSecondaryTextColor,
		autocompleteListTextColor:               Styles.PrimitiveBackgroundColor,
		autocompleteListBackgroundColor:         Styles.MoreContrastBackgroundColor,
		autocompleteListSelectedTextColor:       Styles.PrimitiveBackgroundColor,
		autocompleteListSelectedBackgroundColor: Styles.PrimaryTextColor,
		autocompleteSuggestionTextColor:         Styles.ContrastPrimaryTextColor,
		fieldNoteTextColor:                      Styles.SecondaryTextColor,
		labelColorFocused:                       ColorUnset,
		fieldBackgroundColorFocused:             ColorUnset,
		fieldTextColorFocused:                   ColorUnset,
		placeholderTextColorFocused:             ColorUnset,
	}
}

// SetText sets the current text of the input field.
func (i *InputField) SetText(text string) *InputField {
	i.Lock()

	i.text = text
	i.cursorPos = len(text)
	if i.changed != nil {
		i.Unlock()
		i.changed(text)
	} else {
		i.Unlock()
	}

	return i
}

// GetText returns the current text of the input field.
func (i *InputField) GetText() string {
	i.RLock()
	defer i.RUnlock()

	return i.text
}

// SetLabel sets the text to be displayed before the input area.
func (i *InputField) SetLabel(label string) *InputField {
	i.Lock()
	defer i.Unlock()

	i.label = label
	return i
}

// GetLabel returns the text to be displayed before the input area.
func (i *InputField) GetLabel() string {
	i.RLock()
	defer i.RUnlock()

	return i.label
}

// SetLabelWidth sets the screen width of the label. A value of 0 will cause the
// primitive to use the width of the label string.
func (i *InputField) SetLabelWidth(width int) *InputField {
	i.Lock()
	defer i.Unlock()

	i.labelWidth = width
	return i
}

// SetPlaceholder sets the text to be displayed when the input text is empty.
func (i *InputField) SetPlaceholder(text string) *InputField {
	i.Lock()
	defer i.Unlock()

	i.placeholder = text
	return i
}

// SetLabelColor sets the color of the label.
func (i *InputField) SetLabelColor(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.labelColor = color
	return i
}

// SetLabelColorFocused sets the color of the label when focused.
func (i *InputField) SetLabelColorFocused(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.labelColorFocused = color
	return i
}

// SetFieldBackgroundColor sets the background color of the input area.
func (i *InputField) SetFieldBackgroundColor(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.fieldBackgroundColor = color
	return i
}

// SetFieldBackgroundColorFocused sets the background color of the input area
// when focused.
func (i *InputField) SetFieldBackgroundColorFocused(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.fieldBackgroundColorFocused = color
	return i
}

// SetFieldTextColor sets the text color of the input area.
func (i *InputField) SetFieldTextColor(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.fieldTextColor = color
	return i
}

// SetFieldTextColorFocused sets the text color of the input area when focused.
func (i *InputField) SetFieldTextColorFocused(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.fieldTextColorFocused = color
	return i
}

// SetPlaceholderTextColor sets the text color of placeholder text.
func (i *InputField) SetPlaceholderTextColor(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.placeholderTextColor = color
	return i
}

// SetPlaceholderTextColorFocused sets the text color of placeholder text when
// focused.
func (i *InputField) SetPlaceholderTextColorFocused(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.placeholderTextColorFocused = color
	return i
}

// SetAutocompleteListTextColor sets the text color of the ListItems.
func (i *InputField) SetAutocompleteListTextColor(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.autocompleteListTextColor = color
	return i
}

// SetAutocompleteListBackgroundColor sets the background color of the
// autocomplete list.
func (i *InputField) SetAutocompleteListBackgroundColor(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.autocompleteListBackgroundColor = color
	return i
}

// SetAutocompleteListSelectedTextColor sets the text color of the selected
// ListItem.
func (i *InputField) SetAutocompleteListSelectedTextColor(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.autocompleteListSelectedTextColor = color
	return i
}

// SetAutocompleteListSelectedBackgroundColor sets the background of the
// selected ListItem.
func (i *InputField) SetAutocompleteListSelectedBackgroundColor(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.autocompleteListSelectedBackgroundColor = color
	return i
}

// SetAutocompleteSuggestionTextColor sets the text color of the autocomplete
// suggestion in the input field.
func (i *InputField) SetAutocompleteSuggestionTextColor(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.autocompleteSuggestionTextColor = color
	return i
}

// SetFieldNoteTextColor sets the text color of the note.
func (i *InputField) SetFieldNoteTextColor(color tcell.Color) *InputField {
	i.Lock()
	defer i.Unlock()

	i.fieldNoteTextColor = color
	return i
}

// SetFieldNote sets the text to show below the input field, e.g. when the
// input is invalid.
func (i *InputField) SetFieldNote(note string) *InputField {
	i.Lock()
	defer i.Unlock()

	i.fieldNote = note
	return i
}

// ResetFieldNote sets the note to an empty string.
func (i *InputField) ResetFieldNote() *InputField {
	i.Lock()
	defer i.Unlock()

	i.fieldNote = ""
	return i
}

// SetFieldWidth sets the screen width of the input area. A value of 0 means
// extend as much as possible.
func (i *InputField) SetFieldWidth(width int) *InputField {
	i.Lock()
	defer i.Unlock()

	i.fieldWidth = width
	return i
}

// GetFieldWidth returns this primitive's field width.
func (i *InputField) GetFieldWidth() int {
	i.RLock()
	defer i.RUnlock()

	return i.fieldWidth
}

// GetFieldHeight returns the height of the field.
func (i *InputField) GetFieldHeight() int {
	i.RLock()
	defer i.RUnlock()
	if i.fieldNote == "" {
		return 1
	}
	return 2
}

// GetCursorPosition returns the cursor position.
func (i *InputField) GetCursorPosition() int {
	i.RLock()
	defer i.RUnlock()

	return i.cursorPos
}

// SetCursorPosition sets the cursor position.
func (i *InputField) SetCursorPosition(cursorPos int) {
	i.Lock()
	defer i.Unlock()

	i.cursorPos = cursorPos
}

// SetMaskCharacter sets a character that masks user input on a screen. A value
// of 0 disables masking.
func (i *InputField) SetMaskCharacter(mask rune) *InputField {
	i.Lock()
	defer i.Unlock()

	i.maskCharacter = mask
	return i
}

// SetAutocompleteFunc sets an autocomplete callback function which may return
// ListItems to be selected from a drop-down based on the current text of the
// input field. The drop-down appears only if len(entries) > 0. The callback is
// invoked in this function and whenever the current text changes or when
// Autocomplete() is called. Entries are cleared when the user selects an entry
// or presses Escape.
func (i *InputField) SetAutocompleteFunc(callback func(currentText string) (entries []*ListItem)) *InputField {
	i.Lock()
	i.autocomplete = callback
	i.Unlock()

	i.Autocomplete()
	return i
}

// Autocomplete invokes the autocomplete callback (if there is one). If the
// length of the returned autocomplete entries slice is greater than 0, the
// input field will present the user with a corresponding drop-down list the
// next time the input field is drawn.
//
// It is safe to call this function from any goroutine. Note that the input
// field is not redrawn automatically unless called from the main goroutine
// (e.g. in response to events).
func (i *InputField) Autocomplete() *InputField {
	i.Lock()
	if i.autocomplete == nil {
		i.Unlock()
		return i
	}
	i.Unlock()

	// Do we have any autocomplete entries?
	entries := i.autocomplete(i.text)
	if len(entries) == 0 {
		// No entries, no list.
		i.Lock()
		i.autocompleteList = nil
		i.autocompleteListSuggestion = ""
		i.Unlock()
		return i
	}

	i.Lock()

	// Make a list if we have none.
	if i.autocompleteList == nil {
		i.autocompleteList = NewList()
		i.autocompleteList.
			SetChangedFunc(i.autocompleteChanged).
			ShowSecondaryText(false).
			SetMainTextColor(i.autocompleteListTextColor).
			SetSelectedTextColor(i.autocompleteListSelectedTextColor).
			SetSelectedBackgroundColor(i.autocompleteListSelectedBackgroundColor).
			SetHighlightFullLine(true).
			SetBackgroundColor(i.autocompleteListBackgroundColor)
	}

	// Fill it with the entries.
	currentEntry := -1
	i.autocompleteList.Clear()
	for index, entry := range entries {
		entry.enabled = true
		i.autocompleteList.AddItem(entry)
		if currentEntry < 0 && entry.mainText == i.text {
			currentEntry = index
		}
	}

	// Set the selection if we have one.
	if currentEntry >= 0 {
		i.autocompleteList.SetCurrentItem(currentEntry)
	}

	i.Unlock()
	return i
}

// autocompleteChanged gets called when another item in the
// autocomplete list has been selected.
func (i *InputField) autocompleteChanged(_ int, item *ListItem) {
	if len(item.secondaryText) > 0 && len(i.text) < len(item.secondaryText) {
		i.autocompleteListSuggestion = item.secondaryText[len(i.text):]
	} else if len(item.mainText) > len(i.text)+1 {
		i.autocompleteListSuggestion = item.mainText[len(i.text)+1:]
	} else {
		i.autocompleteListSuggestion = ""
	}
}

// SetAcceptanceFunc sets a handler which may reject the last character that was
// entered (by returning false).
//
// This package defines a number of variables prefixed with InputField which may
// be used for common input (e.g. numbers, maximum text length).
func (i *InputField) SetAcceptanceFunc(handler func(textToCheck string, lastChar rune) bool) *InputField {
	i.Lock()
	defer i.Unlock()

	i.accept = handler
	return i
}

// SetChangedFunc sets a handler which is called whenever the text of the input
// field has changed. It receives the current text (after the change).
func (i *InputField) SetChangedFunc(handler func(text string)) *InputField {
	i.Lock()
	defer i.Unlock()

	i.changed = handler
	return i
}

// SetDoneFunc sets a handler which is called when the user is done entering
// text. The callback function is provided with the key that was pressed, which
// is one of the following:
//
//   - KeyEnter: Done entering text.
//   - KeyEscape: Abort text input.
//   - KeyTab: Move to the next field.
//   - KeyBacktab: Move to the previous field.
func (i *InputField) SetDoneFunc(handler func(key tcell.Key)) *InputField {
	i.Lock()
	defer i.Unlock()

	i.done = handler
	return i
}

// SetFinishedFunc sets a callback invoked when the user leaves this form item.
func (i *InputField) SetFinishedFunc(handler func(key tcell.Key)) *InputField {
	i.Lock()
	defer i.Unlock()

	i.finished = handler
	return i
}

// SetAttributes applies attribute settings to a form item.
func (i *InputField) SetAttributes(attrs *FormItemAttributes) {
	i.SetLabelWidth(attrs.LabelWidth)
	i.SetBackgroundColor(attrs.BackgroundColor)
	i.SetLabelColor(attrs.LabelColor)
	i.SetLabelColorFocused(attrs.LabelColorFocused)
	i.SetFieldTextColor(attrs.FieldTextColor)
	i.SetFieldTextColorFocused(attrs.FieldTextColorFocused)
	i.SetFieldBackgroundColor(attrs.FieldBackgroundColor)
	i.SetFieldBackgroundColorFocused(attrs.FieldBackgroundColorFocused)

	if attrs.FinishedFunc != nil {
		i.SetFinishedFunc(attrs.FinishedFunc)
	}
}

// Draw draws this primitive onto the screen.
func (i *InputField) Draw(screen tcell.Screen) {
	i.Box.Draw(screen)

	i.Lock()
	defer i.Unlock()

	// Select colors
	labelColor := i.labelColor
	fieldBackgroundColor := i.fieldBackgroundColor
	fieldTextColor := i.fieldTextColor
	if i.GetFocusable().HasFocus() {
		if i.labelColorFocused != ColorUnset {
			labelColor = i.labelColorFocused
		}
		if i.fieldBackgroundColorFocused != ColorUnset {
			fieldBackgroundColor = i.fieldBackgroundColorFocused
		}
		if i.fieldTextColorFocused != ColorUnset {
			fieldTextColor = i.fieldTextColorFocused
		}
	}

	// Prepare
	x, y, width, height := i.GetInnerRect()
	rightLimit := x + width
	if height < 1 || rightLimit <= x {
		return
	}

	// Draw label.
	if i.labelWidth > 0 {
		labelWidth := i.labelWidth
		if labelWidth > rightLimit-x {
			labelWidth = rightLimit - x
		}
		Print(screen, i.label, x, y, labelWidth, AlignLeft, labelColor)
		x += labelWidth
	} else {
		_, drawnWidth := Print(screen, i.label, x, y, rightLimit-x, AlignLeft, labelColor)
		x += drawnWidth
	}

	// Draw input area.
	i.fieldX = x
	fieldWidth := i.fieldWidth
	if fieldWidth == 0 {
		fieldWidth = math.MaxInt32
	}
	if rightLimit-x < fieldWidth {
		fieldWidth = rightLimit - x
	}
	fieldStyle := tcell.StyleDefault.Background(fieldBackgroundColor)
	for index := 0; index < fieldWidth; index++ {
		screen.SetContent(x+index, y, ' ', nil, fieldStyle)
	}

	// Text.
	var cursorScreenPos int
	text := i.text
	if text == "" && i.placeholder != "" {
		// Draw placeholder text.
		placeholderTextColor := i.placeholderTextColor
		if i.GetFocusable().HasFocus() && i.placeholderTextColorFocused != ColorUnset {
			placeholderTextColor = i.placeholderTextColorFocused
		}
		Print(screen, Escape(i.placeholder), x, y, fieldWidth, AlignLeft, placeholderTextColor)
		i.offset = 0
	} else {
		// Draw entered text.
		if i.maskCharacter > 0 {
			text = strings.Repeat(string(i.maskCharacter), utf8.RuneCountInString(i.text))
		}
		drawnText := ""
		if fieldWidth >= stringWidth(text) {
			// We have enough space for the full text.
			drawnText = Escape(text)
			Print(screen, drawnText, x, y, fieldWidth, AlignLeft, fieldTextColor)
			i.offset = 0
			iterateString(text, func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
				if textPos >= i.cursorPos {
					return true
				}
				cursorScreenPos += screenWidth
				return false
			})
		} else {
			// The text doesn't fit. Where is the cursor?
			if i.cursorPos < 0 {
				i.cursorPos = 0
			} else if i.cursorPos > len(text) {
				i.cursorPos = len(text)
			}
			// Shift the text so the cursor is inside the field.
			var shiftLeft int
			if i.offset > i.cursorPos {
				i.offset = i.cursorPos
			} else if subWidth := stringWidth(text[i.offset:i.cursorPos]); subWidth > fieldWidth-1 {
				shiftLeft = subWidth - fieldWidth + 1
			}
			currentOffset := i.offset
			iterateString(text, func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
				if textPos >= currentOffset {
					if shiftLeft > 0 {
						i.offset = textPos + textWidth
						shiftLeft -= screenWidth
					} else {
						if textPos+textWidth > i.cursorPos {
							return true
						}
						cursorScreenPos += screenWidth
					}
				}
				return false
			})
			drawnText = Escape(text[i.offset:])
			Print(screen, drawnText, x, y, fieldWidth, AlignLeft, fieldTextColor)
		}
		// Draw suggestion
		if i.maskCharacter == 0 && i.autocompleteListSuggestion != "" {
			Print(screen, i.autocompleteListSuggestion, x+stringWidth(drawnText), y, fieldWidth-stringWidth(drawnText), AlignLeft, i.autocompleteSuggestionTextColor)
		}
	}

	// Draw field note
	if i.fieldNote != "" {
		Print(screen, i.fieldNote, x, y+1, fieldWidth, AlignLeft, i.fieldNoteTextColor)
	}

	// Draw autocomplete list.
	if i.autocompleteList != nil {
		// How much space do we need?
		lheight := i.autocompleteList.GetItemCount()
		lwidth := 0
		for index := 0; index < lheight; index++ {
			entry, _ := i.autocompleteList.GetItemText(index)
			width := TaggedStringWidth(entry)
			if width > lwidth {
				lwidth = width
			}
		}

		// We prefer to drop down but if there is no space, maybe drop up?
		lx := x
		ly := y + 1
		_, sheight := screen.Size()
		if ly+lheight >= sheight && ly-2 > lheight-ly {
			ly = y - lheight
			if ly < 0 {
				ly = 0
			}
		}
		if ly+lheight >= sheight {
			lheight = sheight - ly
		}
		if i.autocompleteList.scrollBarVisibility == ScrollBarAlways || (i.autocompleteList.scrollBarVisibility == ScrollBarAuto && i.autocompleteList.GetItemCount() > lheight) {
			lwidth++ // Add space for scroll bar
		}
		i.autocompleteList.SetRect(lx, ly, lwidth, lheight)
		i.autocompleteList.Draw(screen)
	}

	// Set cursor.
	if i.focus.HasFocus() {
		screen.ShowCursor(x+cursorScreenPos, y)
	}
}

// InputHandler returns the handler for this primitive.
func (i *InputField) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return i.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		i.Lock()

		// Trigger changed events.
		currentText := i.text
		defer func() {
			i.Lock()
			newText := i.text
			i.Unlock()

			if newText != currentText {
				i.Autocomplete()
				if i.changed != nil {
					i.changed(i.text)
				}
			}
		}()

		// Movement functions.
		home := func() { i.cursorPos = 0 }
		end := func() { i.cursorPos = len(i.text) }
		moveLeft := func() {
			iterateStringReverse(i.text[:i.cursorPos], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
				i.cursorPos -= textWidth
				return true
			})
		}
		moveRight := func() {
			iterateString(i.text[i.cursorPos:], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
				i.cursorPos += textWidth
				return true
			})
		}
		moveWordLeft := func() {
			i.cursorPos = len(regexRightWord.ReplaceAllString(i.text[:i.cursorPos], ""))
		}
		moveWordRight := func() {
			i.cursorPos = len(i.text) - len(regexLeftWord.ReplaceAllString(i.text[i.cursorPos:], ""))
		}

		// Add character function. Returns whether or not the rune character is
		// accepted.
		add := func(r rune) bool {
			newText := i.text[:i.cursorPos] + string(r) + i.text[i.cursorPos:]
			if i.accept != nil && !i.accept(newText, r) {
				return false
			}
			i.text = newText
			i.cursorPos += len(string(r))
			return true
		}

		// Finish up.
		finish := func(key tcell.Key) {
			if i.done != nil {
				i.done(key)
			}
			if i.finished != nil {
				i.finished(key)
			}
		}

		// Process key event.
		switch key := event.Key(); key {
		case tcell.KeyRune: // Regular character.
			if event.Modifiers()&tcell.ModAlt > 0 {
				// We accept some Alt- key combinations.
				switch event.Rune() {
				case 'a': // Home.
					home()
				case 'e': // End.
					end()
				case 'b': // Move word left.
					moveWordLeft()
				case 'f': // Move word right.
					moveWordRight()
				default:
					if !add(event.Rune()) {
						i.Unlock()
						return
					}
				}
			} else {
				// Other keys are simply accepted as regular characters.
				if !add(event.Rune()) {
					i.Unlock()
					return
				}
			}
		case tcell.KeyCtrlU: // Delete all.
			i.text = ""
			i.cursorPos = 0
		case tcell.KeyCtrlK: // Delete until the end of the line.
			i.text = i.text[:i.cursorPos]
		case tcell.KeyCtrlW: // Delete last word.
			newText := regexRightWord.ReplaceAllString(i.text[:i.cursorPos], "") + i.text[i.cursorPos:]
			i.cursorPos -= len(i.text) - len(newText)
			i.text = newText
		case tcell.KeyBackspace, tcell.KeyBackspace2: // Delete character before the cursor.
			iterateStringReverse(i.text[:i.cursorPos], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
				i.text = i.text[:textPos] + i.text[textPos+textWidth:]
				i.cursorPos -= textWidth
				return true
			})
			if i.offset >= i.cursorPos {
				i.offset = 0
			}
		case tcell.KeyDelete: // Delete character after the cursor.
			iterateString(i.text[i.cursorPos:], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
				i.text = i.text[:i.cursorPos] + i.text[i.cursorPos+textWidth:]
				return true
			})
		case tcell.KeyLeft:
			if event.Modifiers()&tcell.ModAlt > 0 {
				moveWordLeft()
			} else {
				moveLeft()
			}
		case tcell.KeyRight:
			if event.Modifiers()&tcell.ModAlt > 0 {
				moveWordRight()
			} else {
				moveRight()
			}
		case tcell.KeyHome, tcell.KeyCtrlA:
			home()
		case tcell.KeyEnd, tcell.KeyCtrlE:
			end()
		case tcell.KeyEnter: // We might be done.
			if i.autocompleteList != nil {
				currentItem := i.autocompleteList.GetCurrentItem()
				selectionText := currentItem.mainText
				if currentItem.secondaryText != "" {
					selectionText = currentItem.secondaryText
				}
				i.Unlock()
				i.SetText(selectionText)
				i.Lock()
				i.autocompleteList = nil
				i.autocompleteListSuggestion = ""
			} else {
				finish(key)
			}
			i.Unlock()
			return
		case tcell.KeyEscape:
			if i.autocompleteList != nil {
				i.autocompleteList = nil
				i.autocompleteListSuggestion = ""
			} else {
				finish(key)
			}
			i.Unlock()
			return
		case tcell.KeyDown, tcell.KeyTab: // Autocomplete selection.
			if i.autocompleteList != nil {
				count := i.autocompleteList.GetItemCount()
				newEntry := i.autocompleteList.GetCurrentItemIndex() + 1
				if newEntry >= count {
					newEntry = 0
				}
				i.autocompleteList.SetCurrentItem(newEntry)
				i.Unlock()
			} else {
				i.Unlock()
				finish(key)
			}
			return
		case tcell.KeyUp, tcell.KeyBacktab: // Autocomplete selection.
			if i.autocompleteList != nil {
				newEntry := i.autocompleteList.GetCurrentItemIndex() - 1
				if newEntry < 0 {
					newEntry = i.autocompleteList.GetItemCount() - 1
				}
				i.autocompleteList.SetCurrentItem(newEntry)
				i.Unlock()
			} else {
				i.Unlock()
				finish(key)
			}
			return
		}

		i.Unlock()
	})
}

// MouseHandler returns the mouse handler for this primitive.
func (i *InputField) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return i.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		x, y := event.Position()
		_, rectY, _, _ := i.GetInnerRect()
		if !i.InRect(x, y) {
			return false, nil
		}

		// Process mouse event.
		if action == MouseLeftClick && y == rectY {
			// Determine where to place the cursor.
			if x >= i.fieldX {
				if !iterateString(i.text, func(main rune, comb []rune, textPos int, textWidth int, screenPos int, screenWidth int) bool {
					if x-i.fieldX < screenPos+screenWidth {
						i.cursorPos = textPos
						return true
					}
					return false
				}) {
					i.cursorPos = len(i.text)
				}
			}
			setFocus(i)
			consumed = true
		}

		return
	})
}

var (
	regexRightWord = regexp.MustCompile(`(\w*|\W)$`)
	regexLeftWord  = regexp.MustCompile(`^(\W|\w*)`)
)
