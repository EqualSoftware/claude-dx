package launcher

import (
	"reflect"
	"testing"
)

func TestArgs(t *testing.T) {
	got := Args("/cache/plugin/dev", []string{"--resume", "-p", "hi"})
	want := []string{"--plugin-dir", "/cache/plugin/dev", "--resume", "-p", "hi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Args = %q, want %q", got, want)
	}
}
