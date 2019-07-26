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
		CompletionSpace: false,
		WholeText:       false,
		MaxFields:       5,
	}

	i.Complete.SetBackgroundColor(tcell.ColorWhite)
	i.Complete.SetMainTextColor(tcell.ColorBlack)
	i.Complete.SetHighlightFullLine(true)

	i.SetBackgroundColor(-1)
	i.SetFieldTextColor(-1)
	i.SetChangedFunc(func(text string) {
		if i.ChangeFunc != nil {
			i.ChangeFunc(text)
		}

		if i.Completer == nil {
			return
		}

		if text == "" {
			return
		}

		if !i.WholeText {
			f := strings.Fields(text)
			text = f[len(f)-1]
		}

		completes := i.Completer(text)
		if len(completes) == 0 {
			return
		}

		i.completeMutex.Lock()
		defer i.completeMutex.Unlock()

		i.completes = completes
	})

	return i
}

func (i *Input) Draw(s tcell.Screen) {
	i.completeMutex.RLock()

	var items = make([]*tview.ListItem, len(i.completes))
	for j, c := range i.completes {
		if c.Visual != "" {
			items[j].MainText = c.Visual
		} else {
			items[j].MainText = c.Replace
		}

		items[j].SecondaryText = c.Replace
	}

	i.completeMutex.RUnlock()

	i.Complete.SetItems(items)

	x, y, w, h := i.GetRect()
	i.Complete.SetRect(x, y+1, w, max(i.MaxFields, h))

	i.InputField.Draw(s)
	i.Complete.Draw(s)
}

// SetChangedFunc overrides InputField's SetChangedFunc. This function is called
// when InputField calls its changed func and blocks. This is called before
// completion is done.
func (i *Input) SetChangedFunc(f func(string)) {
	i.ChangeFunc = f
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}
