// Pakcage inputcomplete provides an API for autocompletion on Input Fields.
package inputcomplete

import (
	"strings"
	"sync"

	"github.com/diamondburned/tcell"
	"github.com/diamondburned/tview/v2"
)

type Input struct {
	*tview.InputField

	Complete  *tview.List
	Completer Completer

	// Inserts a space after completion
	CompletionSpace bool

	// Run Completer on the whole field and replace
	WholeText bool

	// TODO: Position enum, top, bottom or auto

	// The max autocompletion field to draw
	MaxFields int

	ChangeFunc func(text string)

	// internals

	completes     []Completion
	completeMutex sync.RWMutex
}

func New() *Input {
	i := &Input{
		InputField:      tview.NewInputField(),
		Complete:        tview.NewList(),
		Completer:       nil,
		CompletionSpace: true,
		WholeText:       false,
		MaxFields:       5,
	}

	i.Complete.SetBackgroundColor(tcell.Color238)
	i.Complete.SetMainTextColor(tcell.Color255)
	i.Complete.SetSelectedBackgroundColor(tcell.Color248)
	i.Complete.SetHighlightFullLine(true)
	i.Complete.ShowSecondaryText(false)
	i.Complete.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		defer tview.Draw()

		switch k, r := ev.Key(), ev.Rune(); {
		case k == tcell.KeyEsc:
			i.setCompletes(nil)
			tview.SetFocus(i.InputField)
			return nil

		case k == tcell.KeyDown, r == 'j':
			c, _ := i.Complete.GetCurrentItem()
			if c == i.Complete.GetItemCount()-1 {
				tview.SetFocus(i.InputField)
				return nil
			}
		case k == tcell.KeyUp, r == 'k':
			c, _ := i.Complete.GetCurrentItem()
			if c == 0 {
				tview.SetFocus(i.InputField)
				return nil
			}
		}

		return ev
	})
	i.Complete.SetSelectedFunc(func(_ int, _, to string, _ rune) {
		text := i.InputField.GetText()
		if i.CompletionSpace {
			to += " "
		}

		/* Overthinking lol
		pos := i.InputField.CursorPos

		var from = 0
		for i := pos - 1; i > 0; i-- {
			if text[i] == ' ' {
				from = i + 1
				break
			}
		}

		text = text[:from] + to + text[pos:]
		i.InputField.SetText(text)
		*/

		// Get the text without the replaced part
		f := strings.Split(text, " ")
		text = strings.Join(f[:len(f)-1], " ")
		if len(f) > 1 {
			text += " "
		}

		// Set the text and set the focus
		i.InputField.SetText(text + to)
		tview.SetFocus(i.InputField)

		// Clear the list after selection
		i.setCompletes(nil)
	})

	i.InputField.SetBackgroundColor(-1)
	i.InputField.SetFieldTextColor(-1)
	i.InputField.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		// TODO: Instead of focusing, try handling keys into the Complete List
		// while still keeping focus onto the InputField
		case tcell.KeyDown:
			i.Complete.SetCurrentItem(0)
			//tview.SetFocus(i.Complete)
			return nil
		case tcell.KeyUp:
			i.Complete.SetCurrentItem(i.Complete.GetItemCount() - 1)
			//tview.SetFocus(i.Complete)
			return nil
		}

		return ev
	})
	i.InputField.SetChangedFunc(func(text string) {
		if i.ChangeFunc != nil {
			i.ChangeFunc(text)
		}

		if i.Completer == nil {
			i.setCompletes(nil)
			return
		}

		if text == "" {
			i.setCompletes(nil)
			return
		}

		if !i.WholeText {
			// Split using space. Any whitespace is not used because the text
			// would be joined using space above.
			f := strings.Split(text, " ")
			text = f[len(f)-1]
		}

		completes := i.Completer(text)
		i.setCompletes(completes)
	})

	return i
}

func (i *Input) Draw(s tcell.Screen) {
	i.completeMutex.RLock()

	var items = make([]*tview.ListItem, 0, len(i.completes))
	for _, c := range i.completes {
		it := &tview.ListItem{}

		if c.Visual != "" {
			it.MainText = c.Visual
		} else {
			it.MainText = c.Replace
		}

		it.SecondaryText = c.Replace
		items = append(items, it)
	}

	i.completeMutex.RUnlock()

	i.Complete.SetItems(items)

	x, y, w, h := i.GetRect()
	i.Complete.SetRect(x, y+1, w, min(len(i.completes), min(i.MaxFields, h)))

	i.InputField.Draw(s)
	i.Complete.Draw(s)
}

// SetChangedFunc overrides InputField's SetChangedFunc. This function is called
// when InputField calls its changed func and blocks. This is called before
// completion is done.
func (i *Input) SetChangedFunc(f func(string)) {
	i.ChangeFunc = f
}

func (i *Input) setCompletes(cs []Completion) {
	i.completeMutex.Lock()
	i.completes = cs
	i.completeMutex.Unlock()
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}
