package main

import (
	"fmt"
	"strings"

	"github.com/perlin-network/wavelet"
	"github.com/perlin-network/wavelet/cmd/cli/tui/forms"
)

func srvCompletion(text string) [][2]string {
	completions := make([][2]string, 0, len(srv.History.Store))

	for _, e := range srv.History.Store {
		if text == "" || strings.Contains(e.ID, text) {
			completions = append(completions, [2]string{
				e.ID, e.String(),
			})
		}
	}

	return completions
}

func getRecipientFormPair(recipient [wavelet.SizeAccountID]byte) forms.Pair {
	return forms.Pair{
		Name: "Recipient",
		Value: func(output string) error {
			if len(output) == wavelet.SizeAccountID {
				copy(recipient[:], []byte(output))
				return nil
			}

			return fmt.Errorf(
				"Invalid recipient length, expected %d, got %d",
				wavelet.SizeAccountID, len(output),
			)
		},
		Validator: forms.ORValidators(
			forms.LetterValidator(), forms.IntValidator(),
		),
		Completer: srvCompletion,
	}
}
