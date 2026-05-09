package executequery_test

import (
	"strings"
	"testing"

	executequery "github.com/apstndb/go-googlesql-executequery"
)

func TestParseFeatureSet(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in       string
		wantBase executequery.FeatureBase
		nEnabled int
		nDisable int
		isErr    bool
	}{
		{in: "", wantBase: executequery.FeatureBaseUnset},
		{in: "NONE", wantBase: executequery.FeatureBaseNone},
		{in: "ALL", wantBase: executequery.FeatureBaseAll},
		{in: "ALL_MINUS_DEV", wantBase: executequery.FeatureBaseAllMinusDev},
		{in: "ALL,-FEATURE_RANGE_TYPE", wantBase: executequery.FeatureBaseAll, nDisable: 1},
		// V_1_1_* form is normalized to match V11* in the Go binding.
		{in: "ALL,-FEATURE_V_1_1_WITH_ON_SUBQUERY", wantBase: executequery.FeatureBaseAll, nDisable: 1},
		{in: "+FEATURE_RANGE_TYPE", wantBase: executequery.FeatureBaseUnset, nEnabled: 1},
		{in: "BOGUS_BASE", isErr: true},
		{in: "ALL,+UNKNOWN_FEATURE_NAME_THAT_DOES_NOT_EXIST", isErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := executequery.ParseFeatureSet(tc.in)
			if tc.isErr {
				if err == nil {
					t.Fatalf("expected error, got %#v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFeatureSet: %v", err)
			}
			if got.Base != tc.wantBase {
				t.Errorf("base: got %v, want %v", got.Base, tc.wantBase)
			}
			if len(got.Enabled) != tc.nEnabled {
				t.Errorf("enabled: got %d, want %d", len(got.Enabled), tc.nEnabled)
			}
			if len(got.Disabled) != tc.nDisable {
				t.Errorf("disabled: got %d, want %d", len(got.Disabled), tc.nDisable)
			}
		})
	}
}

func TestParseRewriteSet(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in       string
		wantBase executequery.RewriteBase
		nEnabled int
		nDisable int
		isErr    bool
	}{
		{in: "", wantBase: executequery.RewriteBaseUnset},
		{in: "NONE", wantBase: executequery.RewriteBaseNone},
		{in: "ALL_MINUS_DEV", wantBase: executequery.RewriteBaseAllMinusDev},
		{in: "DEFAULTS,-FLATTEN", wantBase: executequery.RewriteBaseDefaults, nDisable: 1},
		{in: "DEFAULTS,+ANONYMIZATION", wantBase: executequery.RewriteBaseDefaults, nEnabled: 1},
		{in: "BOGUS", isErr: true},
		{in: "ALL,+REWRITE_DOES_NOT_EXIST", isErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := executequery.ParseRewriteSet(tc.in)
			if tc.isErr {
				if err == nil {
					t.Fatalf("expected error, got %#v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRewriteSet: %v", err)
			}
			if got.Base != tc.wantBase {
				t.Errorf("base: got %v, want %v", got.Base, tc.wantBase)
			}
			if len(got.Enabled) != tc.nEnabled {
				t.Errorf("enabled: got %d, want %d", len(got.Enabled), tc.nEnabled)
			}
			if len(got.Disabled) != tc.nDisable {
				t.Errorf("disabled: got %d, want %d", len(got.Disabled), tc.nDisable)
			}
		})
	}

	// Sanity: the error message for an unknown rewrite should
	// include the user-supplied name verbatim so it's diagnosable.
	if _, err := executequery.ParseRewriteSet("ALL,+REWRITE_BOGUS"); err == nil ||
		!strings.Contains(err.Error(), "REWRITE_BOGUS") {
		t.Errorf("error should mention the bad name, got %v", err)
	}
}
