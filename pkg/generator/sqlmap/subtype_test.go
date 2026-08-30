package sqlmap

import (
	"strings"
	"testing"
)

// A misconfigured subtype hierarchy has to fail at generation time with a
// message naming what is wrong. The alternative is worse than a bad error:
// the generator would emit a schema whose CHECK and composite foreign key
// disagree, so the mistake would surface as every insert failing against a
// live database instead.
func TestGenerate_SubtypeMisconfigurations(t *testing.T) {
	for _, tc := range []struct {
		name  string
		proto string
		// want are fragments the error must mention, so it points at the
		// offending declaration rather than just failing.
		want []string
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
			// subtype_of names the referencing columns but never the
			// referenced ones -- those are always the supertype's primary key
			// -- so the counts have to agree for the composite key to exist.
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
