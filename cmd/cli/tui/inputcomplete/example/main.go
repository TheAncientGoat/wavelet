package main

import (
	"strings"

	"github.com/diamondburned/tview/v2"
	"github.com/perlin-network/wavelet/cmd/cli/tui/inputcomplete"
)

var dict = []string{"uno", "dos", "tres"}

func main() {
	i := inputcomplete.New()
	i.Completer = func(word string) []inputcomplete.Completion {
		cs := make([]inputcomplete.Completion, 0, len(dict))
		for _, d := range dict {
			if strings.HasPrefix(d, word) {
				cs = append(cs, inputcomplete.Completion{
					Visual:  "[red]" + d + "[-],",
					Replace: d,
				})
			}
		}

		return cs
	}

	tview.Initialize()
	tview.SetRoot(i, true)
	tview.SetFocus(i)
	tview.Run()
}
