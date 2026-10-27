package clean

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// Interactive returns the UI of a terminal, made with huh.
func Interactive() UI {
	return UI{Select: selectForm, Confirm: confirmForm}
}

// selectForm shows a multi-select of the enabled items. huh has no disabled option, so the
// disabled items are in a note above the multi-select.
func selectForm(groups []Group) ([]Item, error) {
	var items []Item
	var options []huh.Option[int]
	var notes []string
	for _, g := range groups {
		for _, it := range g.Items {
			if it.Reason != "" {
				notes = append(notes, it.Text()+" "+it.Reason)
				continue
			}
			options = append(options, huh.NewOption(fmt.Sprintf("%s (%s)", it.Text(), FormatSize(it.Size)), len(items)))
			items = append(items, it)
		}
	}
	var fields []huh.Field
	if len(notes) > 0 {
		fields = append(fields, huh.NewNote().Title("Not available").Description(strings.Join(notes, "\n")))
	}
	var picked []int
	if len(options) > 0 {
		fields = append(fields, huh.NewMultiSelect[int]().Title("Select the items to delete").Options(options...).Value(&picked))
	}
	if len(fields) == 0 {
		return nil, nil
	}
	if err := huh.NewForm(huh.NewGroup(fields...)).Run(); err != nil {
		return nil, ignoreAbort(err)
	}
	out := make([]Item, 0, len(picked))
	for _, i := range picked {
		out = append(out, items[i])
	}
	return out, nil
}

// confirmForm asks a question. The default answer is No.
func confirmForm(prompt string) (bool, error) {
	var ok bool
	err := huh.NewConfirm().Title(prompt).Affirmative("Yes").Negative("No").Value(&ok).Run()
	return ok, ignoreAbort(err)
}

// ignoreAbort makes an abort of the person (Ctrl+C) the same as no selection.
func ignoreAbort(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		return nil
	}
	return err
}
