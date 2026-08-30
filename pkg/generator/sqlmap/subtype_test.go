package sqlmap

import (
	"strings"
	"testing"
)

// A misconfigured hierarchy must fail at generation time naming what is wrong.
// Otherwise the CHECK and composite foreign key disagree, and the mistake only
// surfaces as every insert failing against a live database.
func TestGenerate_SubtypeMisconfigurations(t *testing.T) {
	for _, tc := range []struct {
		name  string
		proto string
		want  []string // fragments the error must name
	}{
		{
			name:  "supertype is not a table",
			proto: "subtype_bad_unknown_super.proto",
			want:  []string{"person", "NoSuchThing"},
		},
		{
			name:  "supertype has no subtypes oneof",
			proto: "subtype_bad_no_oneof.proto",
			want:  []string{"person", "Identity", "subtypes option"},
		},
		{
			name:  "supertype's oneof does not name the subtype",
			proto: "subtype_bad_not_named.proto",
			want:  []string{"organisation", "Identity", "does not name it"},
		},
		{
			name:  "a oneof arm is not a table",
			proto: "subtype_bad_arm_not_table.proto",
			want:  []string{"identity", "Organisation", "not a table"},
		},
		{
			// The referenced side is always the supertype's PK, so counts must agree.
			name:  "link columns outnumber the supertype's primary key",
			proto: "subtype_bad_key_count.proto",
			want:  []string{"person", "identity", "2 column(s)", "1 primary key"},
		},
		{
			name:  "subtype_of names a column the table does not have",
			proto: "subtype_bad_unknown_column.proto",
			want:  []string{"person", "nonexistent"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := generateError(t, tc.proto)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("error should mention %q, got: %s", want, got)
				}
			}
		})
	}
}
