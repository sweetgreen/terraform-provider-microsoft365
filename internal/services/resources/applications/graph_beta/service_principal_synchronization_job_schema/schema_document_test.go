package graphBetaServicePrincipalSynchronizationJobSchema

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

const fixtureRuleID = "33333333-3333-3333-3333-333333333333"

func loadFixture(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile("tests/responses/get_schema.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := decodeDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func noSignatures(t *testing.T) func() (map[string]functionSignature, error) {
	return func() (map[string]functionSignature, error) {
		t.Fatal("function definitions were loaded although no expression changed")
		return nil, nil
	}
}

func TestUnitSchemaDocument_Snapshot(t *testing.T) {
	snap, err := snapshot(loadFixture(t), fixtureRuleID)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Version == nil || !strings.HasPrefix(*snap.Version, "Date:2026-03-25") {
		t.Errorf("version = %v", snap.Version)
	}
	if !reflect.DeepEqual(snap.RuleIDs, []string{fixtureRuleID}) {
		t.Errorf("rule IDs = %v", snap.RuleIDs)
	}
	user, ok := snap.Mappings["sys_user"]
	if !ok {
		t.Fatalf("no sys_user mapping in %v", sortedKeys(snap.Mappings))
	}
	if !user.Enabled || user.SourceObjectName != "User" || *user.Name != "Provision Microsoft Entra ID Users" || user.Scope != nil {
		t.Errorf("sys_user = %+v", user)
	}
	if !reflect.DeepEqual(user.FlowTypes, []string{"Add", "Delete", "Update"}) {
		t.Errorf("flow types = %v", user.FlowTypes)
	}
	active := user.AttributeMappings["active"]
	if *active.SourceExpression != `Switch([IsSoftDeleted], , "False", "1", "True", "0")` ||
		*active.DefaultValue != "1" || active.FlowType != "Always" || active.FlowBehavior != "FlowWhenChanged" ||
		active.MatchingPriority != 0 {
		t.Errorf("active = %+v", active)
	}
	if user.AttributeMappings["user_name"].MatchingPriority != 1 {
		t.Errorf("user_name matching priority = %d", user.AttributeMappings["user_name"].MatchingPriority)
	}
	if user.AttributeMappings["email"].DefaultValue != nil {
		t.Errorf("email default value should be null")
	}
	if v := user.AttributeMappings["location"].DefaultValue; v == nil || *v != "" {
		t.Errorf("location default value should be the empty string, got %v", v)
	}
	disabled := snap.Mappings["cmn_cost_center"]
	if disabled.Enabled || disabled.Name != nil || disabled.SourceObjectName != "" {
		t.Errorf("cmn_cost_center = %+v", disabled)
	}

	if _, err := snapshot(loadFixture(t), "nope"); err == nil || !strings.Contains(err.Error(), fixtureRuleID) {
		t.Errorf("unknown rule error = %v", err)
	}
}

// Writing back what was read must send Graph its own document, minus OData annotations.
func TestUnitSchemaDocument_RoundTripIsIdentity(t *testing.T) {
	doc := loadFixture(t)
	snap, err := snapshot(doc, fixtureRuleID)
	if err != nil {
		t.Fatal(err)
	}
	body, err := buildDocument(doc, fixtureRuleID, snap.Mappings, noSignatures(t))
	if err != nil {
		t.Fatal(err)
	}
	want := stripAnnotations(deepCopy(doc))
	if !sameJSON(t, body, want) {
		got, _ := json.MarshalIndent(body, "", " ")
		t.Fatalf("round trip changed the document:\n%s", got)
	}
	for k := range body {
		if strings.Contains(k, "@odata.") {
			t.Errorf("annotation %q was sent", k)
		}
	}
	if _, ok := doc["@odata.context"]; !ok {
		t.Errorf("buildDocument modified its input")
	}
}

func TestUnitSchemaDocument_ChangeOnlyTouchesTheChangedMapping(t *testing.T) {
	doc := loadFixture(t)
	snap, err := snapshot(doc, fixtureRuleID)
	if err != nil {
		t.Fatal(err)
	}
	desired := snap.Mappings
	user := desired["sys_user"]
	active := user.AttributeMappings["active"]
	expr := `Switch([IsSoftDeleted], , "False", "1", "True", "1")`
	active.SourceExpression = &expr
	user.AttributeMappings["active"] = active
	title := "[jobTitle]"
	user.AttributeMappings["title"] = attributeMapping{SourceExpression: &title, FlowType: "Always", FlowBehavior: "FlowWhenChanged"}
	delete(user.AttributeMappings, "location")
	desired["sys_user"] = user
	delete(desired, "cmn_cost_center")

	loads := 0
	body, err := buildDocument(doc, fixtureRuleID, desired, func() (map[string]functionSignature, error) {
		loads++
		return loadSignatures(t), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if loads != 1 {
		t.Errorf("function definitions loaded %d times, want 1", loads)
	}
	if !sameJSON(t, body["directories"], doc["directories"]) {
		t.Errorf("directories changed")
	}
	rule := asMap(asList(body["synchronizationRules"])[0])
	origRule := asMap(asList(doc["synchronizationRules"])[0])
	if !sameJSON(t, rule["metadata"], origRule["metadata"]) {
		t.Errorf("rule metadata changed")
	}
	mappings := asList(rule["objectMappings"])
	var targets []string
	for _, m := range mappings {
		targets = append(targets, stringValue(asMap(m)["targetObjectName"]))
	}
	if !reflect.DeepEqual(targets, []string{"sys_user", "sys_user_group"}) {
		t.Fatalf("object mappings = %v", targets)
	}
	userMap := asMap(mappings[0])
	origUser := asMap(asList(origRule["objectMappings"])[0])
	if !sameJSON(t, userMap["metadata"], origUser["metadata"]) {
		t.Errorf("object mapping metadata changed")
	}
	if !sameJSON(t, mappings[1], asList(origRule["objectMappings"])[1]) {
		t.Errorf("the group mapping changed")
	}
	var names []string
	for _, a := range asList(userMap["attributeMappings"]) {
		am := asMap(a)
		name := stringValue(am["targetAttributeName"])
		names = append(names, name)
		switch name {
		case "active":
			src := asMap(am["source"])
			if src["expression"] != expr {
				t.Errorf("active expression = %v", src["expression"])
			}
			params := asList(src["parameters"])
			if len(params) != 5 || asMap(params[0])["key"] != "source" || asMap(params[4])["key"] != "switchValue" ||
				asMap(asMap(params[4])["value"])["name"] != "1" {
				got, _ := json.Marshal(src)
				t.Errorf("active source = %s", got)
			}
		case "title":
			want := map[string]any{"expression": "[jobTitle]", "name": "jobTitle", "parameters": []any{}, "type": "Attribute"}
			if !sameJSON(t, am["source"], want) || am["defaultValue"] != nil || am["matchingPriority"] != int64(0) {
				got, _ := json.Marshal(am)
				t.Errorf("title = %s", got)
			}
		}
	}
	if want := []string{"active", "email", "user_name", "cost_center", "title"}; !reflect.DeepEqual(names, want) {
		t.Errorf("attribute mappings = %v, want %v", names, want)
	}

	// The document read back after the write must match what was asked for.
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	reread, err := decodeDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	after, err := snapshot(reread, fixtureRuleID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after.Mappings, desired) {
		t.Errorf("read back %+v\nwant %+v", after.Mappings, desired)
	}
}

func TestUnitSchemaDocument_BadExpressionFailsBeforeWriting(t *testing.T) {
	doc := loadFixture(t)
	snap, err := snapshot(doc, fixtureRuleID)
	if err != nil {
		t.Fatal(err)
	}
	user := snap.Mappings["sys_user"]
	active := user.AttributeMappings["active"]
	bad := `switch([IsSoftDeleted], , "False", "1")`
	active.SourceExpression = &bad
	user.AttributeMappings["active"] = active
	_, err = buildDocument(doc, fixtureRuleID, snap.Mappings, func() (map[string]functionSignature, error) {
		return loadSignatures(t), nil
	})
	if err == nil || !strings.Contains(err.Error(), `"sys_user"`) || !strings.Contains(err.Error(), `"active"`) ||
		!strings.Contains(err.Error(), "Switch") {
		t.Errorf("error = %v", err)
	}
}

func TestUnitSchemaDocument_FlowTypesAndScope(t *testing.T) {
	if got := joinFlowTypes([]string{"Delete", "Add", "Update"}); got != "Add,Update,Delete" {
		t.Errorf("joinFlowTypes = %q", got)
	}
	if got := joinFlowTypes([]string{"Update", "Add"}); got != "Add,Update" {
		t.Errorf("joinFlowTypes = %q", got)
	}
	doc := loadFixture(t)
	snap, err := snapshot(doc, fixtureRuleID)
	if err != nil {
		t.Fatal(err)
	}
	group := snap.Mappings["sys_user_group"]
	scope := `{"categoryFilterGroups":[],"groups":[{"clauses":[{"operatorName":"IsTrue","sourceOperandName":"securityEnabled","targetOperand":null}],"name":"security groups"}],"inputFilterGroups":[]}`
	group.Scope = &scope
	group.FlowTypes = []string{"Update", "Add"}
	snap.Mappings["sys_user_group"] = group
	body, err := buildDocument(doc, fixtureRuleID, snap.Mappings, noSignatures(t))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(body)
	reread, err := decodeDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	after, err := snapshot(reread, fixtureRuleID)
	if err != nil {
		t.Fatal(err)
	}
	got := after.Mappings["sys_user_group"]
	if got.Scope == nil || *got.Scope != scope {
		t.Errorf("scope = %v", got.Scope)
	}
	if !reflect.DeepEqual(got.FlowTypes, []string{"Add", "Update"}) {
		t.Errorf("flow types = %v", got.FlowTypes)
	}
}
