package clean

import (
	"errors"
	"fmt"
	"io"
	"slices"
)

// UI is how Run asks the person. The functions in a test need no terminal.
type UI struct {
	// Select shows the groups and returns the items that the person selects.
	Select func(groups []Group) ([]Item, error)
	// Confirm asks the question. The default answer is No.
	Confirm func(prompt string) (bool, error)
}

// Run shows the items, asks the person to select and confirm, and deletes the selected items. It
// deletes nothing without a selection and a Yes. It continues after an error and returns all errors.
func Run(w io.Writer, cacheDir, own string, ui UI) error {
	groups, err := Scan(cacheDir, own)
	if err != nil {
		return err
	}
	groups = selectable(groups)
	if !hasEnabled(groups) {
		fmt.Fprintln(w, "No data to delete.")
		return nil
	}
	picked, err := ui.Select(groups)
	if err != nil {
		return err
	}
	picked = enabled(picked)
	if len(picked) == 0 {
		fmt.Fprintln(w, "Nothing deleted.")
		return nil
	}
	for _, it := range picked {
		fmt.Fprintf(w, "  %s (%s)\n", it.Text(), FormatSize(it.Size))
	}
	ok, err := ui.Confirm(fmt.Sprintf("Delete %d items (%s)?", len(picked), FormatSize(total(picked))))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(w, "Nothing deleted.")
		return nil
	}
	return remove(w, picked)
}

// remove deletes the items and prints the result. It counts an item if it removed a path.
func remove(w io.Writer, items []Item) error {
	var errs []error
	var res result
	count := 0
	for _, it := range items {
		r, err := it.remove()
		res.skipped += r.skipped
		res.bytes += r.bytes
		if r.removed > 0 {
			count++
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", it.Text(), err))
		}
	}
	fmt.Fprintf(w, "Deleted %d items (%s).\n", count, FormatSize(res.bytes))
	if res.skipped > 0 {
		fmt.Fprintf(w, "Skipped %d tasks that changed after the selection.\n", res.skipped)
	}
	return errors.Join(errs...)
}

// hasEnabled reports whether a person can select an item in groups.
func hasEnabled(groups []Group) bool {
	for _, g := range groups {
		if len(enabled(g.Items)) > 0 {
			return true
		}
	}
	return false
}

// selectable returns the groups that a person can select from.
func selectable(groups []Group) []Group {
	return slices.DeleteFunc(slices.Clone(groups), func(g Group) bool { return g.ListOnly })
}

// List prints the groups with the sizes and the status of each item. It deletes nothing.
func List(w io.Writer, groups []Group) {
	for _, g := range groups {
		fmt.Fprintln(w, g.Name)
		for _, it := range g.Items {
			line := fmt.Sprintf("  %s  %s", it.Label, FormatSize(it.Size))
			if it.Reason != "" {
				line += "  " + it.Reason
			}
			fmt.Fprintln(w, line)
		}
	}
}

// FormatSize returns size with the unit B, KB, MB or GB (base 1024).
func FormatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	v := float64(size)
	for _, unit := range []string{"KB", "MB", "GB"} {
		v /= 1024
		if v < 1024 || unit == "GB" {
			return fmt.Sprintf("%.1f %s", v, unit)
		}
	}
	return ""
}

// enabled returns the items that a person can delete.
func enabled(items []Item) []Item {
	var out []Item
	for _, it := range items {
		if it.Reason == "" && it.remove != nil {
			out = append(out, it)
		}
	}
	return out
}

func total(items []Item) int64 {
	var sum int64
	for _, it := range items {
		sum += it.Size
	}
	return sum
}
