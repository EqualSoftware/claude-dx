package result

import "testing"

func TestTable(t *testing.T) {
	got := Table([][]string{
		{"ADR-001", "Accepted", "Load dx with launch flags"},
		{"ADR-012", "Proposed", "Use SemVer"},
	})
	want := "ADR-001  Accepted  Load dx with launch flags\nADR-012  Proposed  Use SemVer\n"
	if got != want {
		t.Fatalf("Table =\n%q\nwant\n%q", got, want)
	}
}

// TestTableRunes is AC-1 of TASK-045: a cell with "é" does not move the next column.
func TestTableRunes(t *testing.T) {
	got := Table([][]string{{"é", "a"}, {"ab", "b"}})
	want := "é   a\nab  b\n"
	if got != want {
		t.Fatalf("Table =\n%q\nwant\n%q", got, want)
	}
}

func TestErrorf(t *testing.T) {
	r := Errorf("ADR-%03d does not exist.", 7)
	if !r.IsError || len(r.Content) != 1 {
		t.Fatalf("Errorf = %+v", r)
	}
}
