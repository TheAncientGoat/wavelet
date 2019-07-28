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

	// the index to draw the completer with X offset
	completerIndex int
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

	i.InputField.SetBackgroundColor(-1)
	i.InputField.SetFieldTextColor(-1)
	i.InputField.SetInputCapture(i.inputBinds)
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

			i.completerIndex = len(strings.Join(f[:len(f)-1], " "))
			if len(f) > 1 {
				i.completerIndex++
			}
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

	if i.Complete.GetCurrentItem() < 0 {
		i.Complete.SetCurrentItem(0)
	}

	x, y, w, h := i.GetRect()

	// The positions for the autocompleter
	var cX, cY = x + i.completerIndex, y + 1

	// Calculate the height for the autocompleter
	height := min(len(items), min(i.MaxFields, h))
	width := w - i.completerIndex

	//panic(fmt.Sprint("bruh", height, len(items) + 1, i.MaxFields, h))

	// If the drop-down completer goes out of screen
	if cY+height > y+h {
		// Just drop up lol
		cY = y - height
	}

	i.Complete.SetRect(cX, cY, width, height)

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
