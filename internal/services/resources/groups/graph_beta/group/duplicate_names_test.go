package graphBetaGroup

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDisplayNameFilter_EscapesSingleQuotes(t *testing.T) {
	cases := map[string]string{
		"POS Ring 1":          "displayName eq 'POS Ring 1'",
		"Joe's Devices":       "displayName eq 'Joe''s Devices'",
		"''":                  "displayName eq ''''''",
		"A & B #1 + C ? 100%": "displayName eq 'A & B #1 + C ? 100%'",
	}
	for name, want := range cases {
		if got := displayNameFilter(name); got != want {
			t.Errorf("displayNameFilter(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestDuplicateNameDetail_NamesEveryID(t *testing.T) {
	one := duplicateNameDetail("POS Ring 1", []string{"11111111-1111-1111-1111-111111111111"})
	for _, want := range []string{"1 group(s)", `"POS Ring 1"`, "11111111-1111-1111-1111-111111111111", "import", "prevent_duplicate_names = false"} {
		if !strings.Contains(one, want) {
			t.Errorf("one-match detail %q is missing %q", one, want)
		}
	}

	two := duplicateNameDetail("POS Ring 1", []string{"11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"})
	for _, want := range []string{"2 group(s)", "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"} {
		if !strings.Contains(two, want) {
			t.Errorf("two-match detail %q is missing %q", two, want)
		}
	}
}

func TestShouldCheckDuplicateName(t *testing.T) {
	cases := []struct {
		name string
		flag types.Bool
		dn   types.String
		want bool
	}{
		{"flag on, name known", types.BoolValue(true), types.StringValue("G"), true},
		{"flag off", types.BoolValue(false), types.StringValue("G"), false},
		{"flag null", types.BoolNull(), types.StringValue("G"), false},
		{"flag unknown", types.BoolUnknown(), types.StringValue("G"), false},
		{"name unknown", types.BoolValue(true), types.StringUnknown(), false},
	}
	for _, c := range cases {
		m := &GroupResourceModel{PreventDuplicateNames: c.flag, DisplayName: c.dn}
		if got := shouldCheckDuplicateName(m); got != c.want {
			t.Errorf("%s: shouldCheckDuplicateName = %t, want %t", c.name, got, c.want)
		}
	}
}

// Update skips the PATCH when the body it would send already matches state. A change to a
// provider-side attribute alone must count as no Graph change; a description change must not.
func TestUpdateBodyMatchesState_OnlyForProviderSideChanges(t *testing.T) {
	ctx := context.Background()
	state := consistencyTestModel("(device.displayName -startsWith \"POS\")", "DynamicMembership")
	state.Description = types.StringValue("before")
	state.HardDelete = types.BoolValue(false)
	state.PreventDuplicateNames = types.BoolValue(false)

	plan := *state
	plan.HardDelete = types.BoolValue(true)
	plan.PreventDuplicateNames = types.BoolValue(true)
	if !newGroupConsistency(ctx, sentUpdate(t, &plan)).matches(state) {
		t.Error("a provider-side-only change should not need a PATCH")
	}

	plan.Description = types.StringValue("after")
	if newGroupConsistency(ctx, sentUpdate(t, &plan)).matches(state) {
		t.Error("a description change must be sent")
	}
}
