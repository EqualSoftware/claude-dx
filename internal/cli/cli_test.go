package cli

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		args []string
		want Command
	}{
		{nil, Command{}},
		{[]string{"--resume"}, Command{Args: []string{"--resume"}}},
		{[]string{"-p", "hi"}, Command{Args: []string{"-p", "hi"}}},
		{[]string{"--help"}, Command{Name: "help", Args: []string{}}},
		{[]string{"-h"}, Command{Name: "help", Args: []string{}}},
		{[]string{"--version"}, Command{Name: "version", Args: []string{}}},
		{[]string{"clean", "--list"}, Command{Name: "clean", Args: []string{"--list"}}},
		{[]string{"cache", "clean"}, Command{Args: []string{"cache", "clean"}}},
		{[]string{"--", "--help"}, Command{Args: []string{"--help"}}},
		{[]string{"--", "update"}, Command{Args: []string{"update"}}},
		{[]string{"--"}, Command{Args: []string{}}},
		{[]string{"--resume", "--help"}, Command{Args: []string{"--resume", "--help"}}},
	}
	for _, tt := range tests {
		if got := Parse(tt.args); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Parse(%q) = %+v, want %+v", tt.args, got, tt.want)
		}
	}
}
