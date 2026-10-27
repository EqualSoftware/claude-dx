package spec

import (
	"errors"
	"strings"

	"github.com/EqualSoftware/claude-dx/internal/markdown"
)

// Update holds the new frontmatter fields. An empty string or a nil list keeps the field. An empty
// list removes all items. The slug and the status are not fields of Update.
type Update struct {
	Title     string
	Milestone string
	Priority  string
	Area      []string
	Related   []string
}

// changed returns the names of the fields that u sets, in the order of the frontmatter.
func (u Update) changed() []string {
	var names []string
	for _, f := range []struct {
		name string
		set  bool
	}{
		{"title", u.Title != ""}, {"milestone", u.Milestone != ""}, {"priority", u.Priority != ""},
		{"area", u.Area != nil}, {"related", u.Related != nil},
	} {
		if f.set {
			names = append(names, f.name)
		}
	}
	return names
}

// checkUpdate returns an error for an invalid field, or for an Update with no field.
func (s Store) checkUpdate(slug string, u Update) error {
	if len(u.changed()) == 0 {
		return errors.New("give a field to change")
	}
	if u.Title != "" {
		if err := s.checkTitle(u.Title, slug); err != nil {
			return err
		}
	}
	if u.Milestone != "" {
		if err := checkMilestone(u.Milestone); err != nil {
			return err
		}
	}
	if u.Priority != "" {
		if err := oneOf("priority", u.Priority, Priorities); err != nil {
			return err
		}
	}
	return s.checkRelated(u.Related)
}

// UpdateFields changes the frontmatter fields of the spec slug and adds one Document History row. It
// does not change the status. It validates all fields before it changes one.
func (s Store) UpdateFields(slug string, u Update) error {
	return s.edit(slug, func(d *doc) error {
		if err := s.checkUpdate(slug, u); err != nil {
			return err
		}
		u.apply(d)
		return s.addHistory(d, "Changed fields: "+strings.Join(u.changed(), ", ")+".")
	})
}

// apply sets the fields of u in d.
func (u Update) apply(d *doc) {
	if u.Title != "" {
		title := strings.TrimSpace(u.Title)
		d.front = markdown.SetField(d.front, "title", title)
		d.body = markdown.ReplaceTitle(d.body, "# "+title)
	}
	if u.Milestone != "" {
		d.front = markdown.SetField(d.front, "milestone", u.Milestone)
	}
	if u.Priority != "" {
		d.front = markdown.SetField(d.front, "priority", u.Priority)
	}
	if u.Area != nil {
		d.front = markdown.SetField(d.front, "area", u.Area)
	}
	if u.Related != nil {
		d.front = markdown.SetField(d.front, "related", u.Related)
	}
}
