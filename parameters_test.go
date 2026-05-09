package executequery_test

import (
	"testing"

	googlesql "github.com/goccy/go-googlesql"

	executequery "github.com/apstndb/go-googlesql-executequery"
)

func TestParseParameters(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in    string
		want  []executequery.QueryParameter
		isErr bool
	}{
		{in: "", want: nil},
		{in: "a=1", want: []executequery.QueryParameter{{Name: "a", Type: googlesql.TypeKindTypeInt64, Literal: "1"}}},
		{in: "a=3.14", want: []executequery.QueryParameter{{Name: "a", Type: googlesql.TypeKindTypeDouble, Literal: "3.14"}}},
		{in: "a='x'", want: []executequery.QueryParameter{{Name: "a", Type: googlesql.TypeKindTypeString, Literal: "'x'"}}},
		{in: "b=TRUE,c=FALSE", want: []executequery.QueryParameter{
			{Name: "b", Type: googlesql.TypeKindTypeBool, Literal: "TRUE"},
			{Name: "c", Type: googlesql.TypeKindTypeBool, Literal: "FALSE"},
		}},
		{in: "noeq", isErr: true},
		{in: "=1", isErr: true},
		{in: "a=NULL", isErr: true},
		{in: "a=foo", isErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := executequery.ParseParameters(tc.in)
			if tc.isErr {
				if err == nil {
					t.Fatalf("expected error, got %#v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseParameters: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %d entries, want %d (%#v)", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("entry %d: got %#v, want %#v", i, got[i], tc.want[i])
				}
			}
		})
	}
}
