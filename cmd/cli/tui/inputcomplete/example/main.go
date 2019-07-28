package main

import (
	"strings"

	"github.com/diamondburned/tcell"
	"github.com/diamondburned/tview/v2"
	"github.com/perlin-network/wavelet/cmd/cli/tui/clearbg"
	"github.com/perlin-network/wavelet/cmd/cli/tui/inputcomplete"
)

var dict = []string{
	"uno", "dos", "tres", "cuatro", "cinco",
	"seis", "siete", "ocho", "nueve", "diez",
}

func main() {
	clearbg.Enable()

	i := inputcomplete.New()
	i.Completer = func(word string) []inputcomplete.Completion {
		cs := make([]inputcomplete.Completion, 0, len(dict))
		for _, d := range dict {
			if strings.HasPrefix(d, word) {
				cs = append(cs, inputcomplete.Completion{
					Visual:  "[white]" + d + "[-]",
					Replace: d,
				})
			}
		}

		return cs
	}

	tv := tview.NewTextView()
	tv.SetBackgroundColor(tcell.ColorBlack)
	tv.SetText("Type the first 10 numbers in Spanish:\n" +
		strings.Join(dict, " "))

	f := tview.NewFlex()
	f.SetDirection(tview.FlexRow)
	f.AddItem(i, 1, 1, true)
	f.AddItem(tv, 0, 1, false)

	tview.Initialize()
	tview.SetRoot(f, true)
	tview.SetFocus(f)
	tview.Run()
}
