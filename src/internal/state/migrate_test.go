package state

import "testing"

func TestMigrationVersion(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		file string
		want int
		ok   bool
	}{
		{"standard", "0001_init.sql", 1, true},
		{"four digit", "1234_add_table.sql", 1234, true},
		{"no underscore", "0001.sql", 0, false},
		{"empty prefix", "_x.sql", 0, false},
		{"non numeric", "abcd_x.sql", 0, false},
		{"suffix only", "foo.sql", 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := migrationVersion(tc.file)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("migrationVersion(%q) = (%d, %v), want (%d, %v)", tc.file, got, ok, tc.want, tc.ok)
			}
		})
	}
}
