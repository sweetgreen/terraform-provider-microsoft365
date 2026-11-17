package graphBetaServicePrincipalSynchronizationJobSchema

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func loadSignatures(t *testing.T) map[string]functionSignature {
	t.Helper()
	raw, err := os.ReadFile("tests/responses/get_functions.json")
	if err != nil {
		t.Fatal(err)
	}
	var list struct {
		Value []functionSignature `json:"value"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		t.Fatal(err)
	}
	out := map[string]functionSignature{}
	for _, s := range list.Value {
		out[s.ID] = s
	}
	return out
}

func TestUnitExpression_Canonical(t *testing.T) {
	cases := map[string]string{
		`[mail]`:       `[mail]`,
		` [mail] `:     `[mail]`,
		`"1"`:          `"1"`,
		`"say \"hi\""`: `"say \"hi\""`,
		`Switch([IsSoftDeleted], , "False", "1", "True", "0")`:  `Switch([IsSoftDeleted], , "False", "1", "True", "0")`,
		`Switch([IsSoftDeleted],,"False","1","True","0")`:       `Switch([IsSoftDeleted], , "False", "1", "True", "0")`,
		`Join(" ", [givenName], [surname])`:                     `Join(" ", [givenName], [surname])`,
		`IIF(IsNullOrEmpty([mail]),[userPrincipalName],[mail])`: `IIF(IsNullOrEmpty([mail]), [userPrincipalName], [mail])`,
		`Guid()`:                      `Guid()`,
		`[extension_0123_costCenter]`: `[extension_0123_costCenter]`,
	}
	for in, want := range cases {
		got, err := canonicalExpression(in)
		if err != nil {
			t.Errorf("canonicalExpression(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("canonicalExpression(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUnitExpression_Errors(t *testing.T) {
	cases := map[string]string{
		``:                "empty",
		`[mail`:           "closing ]",
		`[]`:              "no name",
		`"abc`:            "closing double quote",
		`Switch([a], "b"`: "closing )",
		`Switch [a]`:      "expected (",
		`[a] [b]`:         "after the end",
		`Switch([a] "b")`: "expected , or )",
		`1`:               "unexpected",
	}
	for in, want := range cases {
		_, err := parseExpression(in)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("parseExpression(%q) error = %v, want it to mention %q", in, err, want)
		}
	}
}

// The tree built from an expression must match the tree Graph returned for it.
func TestUnitExpression_ToSourceMatchesGraph(t *testing.T) {
	raw, err := os.ReadFile("tests/responses/get_schema.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := decodeDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	sigs := loadSignatures(t)
	checked := 0
	for _, r := range asList(doc["synchronizationRules"]) {
		for _, om := range asList(asMap(r)["objectMappings"]) {
			for _, am := range asList(asMap(om)["attributeMappings"]) {
				src := asMap(asMap(am)["source"])
				if src == nil {
					continue
				}
				expr := src["expression"].(string)
				node, err := parseExpression(expr)
				if err != nil {
					t.Fatalf("parse %q: %v", expr, err)
				}
				if node.canonical() != expr {
					t.Errorf("canonical(%q) = %q", expr, node.canonical())
				}
				built, err := node.toSource(sigs)
				if err != nil {
					t.Fatalf("toSource %q: %v", expr, err)
				}
				if !sameJSON(t, built, src) {
					got, _ := json.Marshal(built)
					want, _ := json.Marshal(src)
					t.Errorf("tree for %q:\n got %s\nwant %s", expr, got, want)
				}
				checked++
			}
		}
	}
	if checked < 10 {
		t.Fatalf("checked only %d sources", checked)
	}
}

func TestUnitExpression_ArgumentKeys(t *testing.T) {
	sigs := loadSignatures(t)
	node, err := parseExpression(`Join(", ", [a], [b], [c])`)
	if err != nil {
		t.Fatal(err)
	}
	src, err := node.toSource(sigs)
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, p := range src["parameters"].([]any) {
		keys = append(keys, p.(map[string]any)["key"].(string))
	}
	if want := []string{"separator", "source", "source", "source"}; !reflect.DeepEqual(keys, want) {
		t.Errorf("Join keys = %v, want %v", keys, want)
	}

	for expr, want := range map[string]string{
		`switch([a], , "x", "y")`: `did you mean "Switch"`,
		`Nope([a])`:               `unknown function "Nope"`,
		`Not([a], [b])`:           "at most 1 arguments",
		`Switch(, "x", "y")`:      `missing its required argument "source"`,
		`Append([a])`:             `missing its required argument "suffix"`,
	} {
		node, err := parseExpression(expr)
		if err != nil {
			t.Fatalf("parse %q: %v", expr, err)
		}
		if _, err := node.toSource(sigs); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("toSource(%q) error = %v, want it to mention %q", expr, err, want)
		}
	}
}

func sameJSON(t *testing.T, a, b any) bool {
	t.Helper()
	norm := func(v any) any {
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		var out any
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	return reflect.DeepEqual(norm(a), norm(b))
}
