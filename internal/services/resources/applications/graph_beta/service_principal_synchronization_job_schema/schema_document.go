package graphBetaServicePrincipalSynchronizationJobSchema

// This file has no dependencies outside the standard library so the read and write
// logic can be tested against real schema documents without compiling the Graph SDK.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// flowTypeOrder is the order Graph uses when it writes objectMapping.flowTypes.
var flowTypeOrder = []string{"Add", "Update", "Delete"}

// objectMapping is the managed part of one objectMapping, keyed by targetObjectName.
type objectMapping struct {
	Name             *string
	Enabled          bool
	SourceObjectName string
	FlowTypes        []string
	// Scope is the compact JSON of the scoping filter, or nil when there is none.
	Scope             *string
	AttributeMappings map[string]attributeMapping
}

// attributeMapping is the managed part of one attributeMapping, keyed by targetAttributeName.
type attributeMapping struct {
	// SourceExpression is source.expression, or nil when the mapping has no source.
	SourceExpression        *string
	DefaultValue            *string
	FlowType                string
	FlowBehavior            string
	MatchingPriority        int64
	ExportMissingReferences bool
}

// schemaSnapshot is what Read keeps from GET .../schema.
type schemaSnapshot struct {
	Version  *string
	RuleIDs  []string
	Mappings map[string]objectMapping
}

// decodeDocument decodes a schema document, keeping numbers exact.
func decodeDocument(raw []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var doc map[string]any
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode synchronization schema: %w", err)
	}
	if doc == nil {
		return nil, fmt.Errorf("synchronization schema response is empty")
	}
	return doc, nil
}

func ruleIDs(doc map[string]any) []string {
	var ids []string
	for _, r := range asList(doc["synchronizationRules"]) {
		if id, ok := asMap(r)["id"].(string); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func findRule(doc map[string]any, ruleID string) (map[string]any, error) {
	for _, r := range asList(doc["synchronizationRules"]) {
		rule := asMap(r)
		if id, _ := rule["id"].(string); id == ruleID {
			return rule, nil
		}
	}
	return nil, fmt.Errorf("synchronization rule %q not found; the schema has rules %s", ruleID, quoteList(ruleIDs(doc)))
}

// snapshot extracts the managed fields of one rule.
func snapshot(doc map[string]any, ruleID string) (*schemaSnapshot, error) {
	snap := &schemaSnapshot{RuleIDs: ruleIDs(doc), Mappings: map[string]objectMapping{}}
	if v, ok := doc["version"].(string); ok {
		snap.Version = &v
	}
	rule, err := findRule(doc, ruleID)
	if err != nil {
		return nil, err
	}
	for _, om := range asList(rule["objectMappings"]) {
		m := asMap(om)
		target := stringValue(m["targetObjectName"])
		if _, dup := snap.Mappings[target]; dup {
			return nil, fmt.Errorf("rule %q has more than one object mapping with targetObjectName %q; this resource keys object mappings by target object and cannot manage it", ruleID, target)
		}
		mapping := objectMapping{
			Name:              optionalString(m["name"]),
			Enabled:           m["enabled"] == true,
			SourceObjectName:  stringValue(m["sourceObjectName"]),
			FlowTypes:         splitFlowTypes(stringValue(m["flowTypes"])),
			AttributeMappings: map[string]attributeMapping{},
		}
		if scope, ok := m["scope"]; ok && scope != nil {
			b, err := json.Marshal(scope)
			if err != nil {
				return nil, fmt.Errorf("encode scope of %q: %w", target, err)
			}
			s := string(b)
			mapping.Scope = &s
		}
		for _, am := range asList(m["attributeMappings"]) {
			a := asMap(am)
			name := stringValue(a["targetAttributeName"])
			if _, dup := mapping.AttributeMappings[name]; dup {
				return nil, fmt.Errorf("object mapping %q has more than one attribute mapping for %q; this resource keys attribute mappings by target attribute and cannot manage it", target, name)
			}
			attr := attributeMapping{
				DefaultValue:            optionalString(a["defaultValue"]),
				FlowType:                stringValue(a["flowType"]),
				FlowBehavior:            stringValue(a["flowBehavior"]),
				ExportMissingReferences: a["exportMissingReferences"] == true,
			}
			if n, ok := a["matchingPriority"].(json.Number); ok {
				if attr.MatchingPriority, err = n.Int64(); err != nil {
					return nil, fmt.Errorf("matchingPriority of %s.%s: %w", target, name, err)
				}
			}
			if src := asMap(a["source"]); src != nil {
				attr.SourceExpression = optionalString(src["expression"])
			}
			mapping.AttributeMappings[name] = attr
		}
		snap.Mappings[target] = mapping
	}
	return snap, nil
}

// buildDocument returns the body for PUT .../schema. It starts from the document Graph
// returned and replaces only the object mappings of one rule, so directories, other
// rules, rule settings and object mapping metadata are sent back unchanged.
//
// Object and attribute mappings that exist in Graph keep their position; ones that are
// not in the desired configuration are removed; new ones are appended in key order. An
// attribute mapping whose expression is unchanged keeps the source tree Graph returned.
// Changed expressions are parsed and built with the function signatures from
// loadSignatures, which is only called when one is needed.
func buildDocument(
	doc map[string]any,
	ruleID string,
	desired map[string]objectMapping,
	loadSignatures func() (map[string]functionSignature, error),
) (map[string]any, error) {
	out, ok := stripAnnotations(deepCopy(doc)).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("synchronization schema is not a JSON object")
	}
	rule, err := findRule(out, ruleID)
	if err != nil {
		return nil, err
	}
	var signatures map[string]functionSignature
	sourceFor := func(expression string) (map[string]any, error) {
		node, err := parseExpression(expression)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", expression, err)
		}
		if signatures == nil && containsFunction(node) {
			if signatures, err = loadSignatures(); err != nil {
				return nil, fmt.Errorf("load expression function definitions: %w", err)
			}
		}
		src, err := node.toSource(signatures)
		if err != nil {
			return nil, fmt.Errorf("expression %q: %w", expression, err)
		}
		return src, nil
	}

	var mappings []any
	used := map[string]bool{}
	for _, om := range asList(rule["objectMappings"]) {
		existing := asMap(om)
		target := stringValue(existing["targetObjectName"])
		want, ok := desired[target]
		if !ok || used[target] {
			continue
		}
		used[target] = true
		built, err := buildObjectMapping(existing, target, want, sourceFor)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, built)
	}
	for _, target := range sortedKeys(desired) {
		if used[target] {
			continue
		}
		built, err := buildObjectMapping(nil, target, desired[target], sourceFor)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, built)
	}
	if mappings == nil {
		mappings = []any{}
	}
	rule["objectMappings"] = mappings
	return out, nil
}

func buildObjectMapping(
	existing map[string]any,
	target string,
	want objectMapping,
	sourceFor func(string) (map[string]any, error),
) (map[string]any, error) {
	out := map[string]any{}
	for k, v := range existing {
		out[k] = v
	}
	out["targetObjectName"] = target
	out["name"] = nullableString(want.Name)
	out["enabled"] = want.Enabled
	out["sourceObjectName"] = want.SourceObjectName
	out["flowTypes"] = joinFlowTypes(want.FlowTypes)
	out["scope"] = nil
	if want.Scope != nil {
		dec := json.NewDecoder(strings.NewReader(*want.Scope))
		dec.UseNumber()
		var scope any
		if err := dec.Decode(&scope); err != nil {
			return nil, fmt.Errorf("scope of object mapping %q is not valid JSON: %w", target, err)
		}
		out["scope"] = scope
	}

	var attrs []any
	used := map[string]bool{}
	for _, am := range asList(existing["attributeMappings"]) {
		current := asMap(am)
		name := stringValue(current["targetAttributeName"])
		want, ok := want.AttributeMappings[name]
		if !ok || used[name] {
			continue
		}
		used[name] = true
		built, err := buildAttributeMapping(current, name, want, sourceFor)
		if err != nil {
			return nil, fmt.Errorf("object mapping %q: %w", target, err)
		}
		attrs = append(attrs, built)
	}
	for _, name := range sortedKeys(want.AttributeMappings) {
		if used[name] {
			continue
		}
		built, err := buildAttributeMapping(nil, name, want.AttributeMappings[name], sourceFor)
		if err != nil {
			return nil, fmt.Errorf("object mapping %q: %w", target, err)
		}
		attrs = append(attrs, built)
	}
	if attrs == nil {
		attrs = []any{}
	}
	out["attributeMappings"] = attrs
	return out, nil
}

func buildAttributeMapping(
	existing map[string]any,
	name string,
	want attributeMapping,
	sourceFor func(string) (map[string]any, error),
) (map[string]any, error) {
	out := map[string]any{}
	for k, v := range existing {
		out[k] = v
	}
	out["targetAttributeName"] = name
	out["defaultValue"] = nullableString(want.DefaultValue)
	out["flowType"] = want.FlowType
	out["flowBehavior"] = want.FlowBehavior
	out["matchingPriority"] = want.MatchingPriority
	out["exportMissingReferences"] = want.ExportMissingReferences
	switch {
	case want.SourceExpression == nil:
		out["source"] = nil
	case existingExpression(existing) == *want.SourceExpression:
		// Unchanged: send back the tree Graph returned.
	default:
		src, err := sourceFor(*want.SourceExpression)
		if err != nil {
			return nil, fmt.Errorf("attribute mapping %q: %w", name, err)
		}
		out["source"] = src
	}
	return out, nil
}

func existingExpression(existing map[string]any) string {
	src := asMap(existing["source"])
	if src == nil {
		return "\x00no source"
	}
	s, ok := src["expression"].(string)
	if !ok {
		return "\x00no expression"
	}
	return s
}

// containsFunction reports whether building the source needs function definitions.
// Attributes and constants have no arguments, so only the top node needs checking.
func containsFunction(n *exprNode) bool {
	return n != nil && n.Type == sourceTypeFunction
}

// stripAnnotations removes OData annotations such as @odata.context, which Graph adds
// to responses and does not accept back.
func stripAnnotations(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			if strings.Contains(k, "@odata.") {
				delete(t, k)
				continue
			}
			t[k] = stripAnnotations(child)
		}
		return t
	case []any:
		for i, child := range t {
			t[i] = stripAnnotations(child)
		}
		return t
	default:
		return v
	}
}

func deepCopy(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, child := range t {
			out[k] = deepCopy(child)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, child := range t {
			out[i] = deepCopy(child)
		}
		return out
	default:
		return v
	}
}

func splitFlowTypes(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

func joinFlowTypes(types []string) string {
	want := map[string]bool{}
	for _, t := range types {
		want[t] = true
	}
	var out []string
	for _, t := range flowTypeOrder {
		if want[t] {
			out = append(out, t)
			delete(want, t)
		}
	}
	// Values Graph may add later are kept, after the known ones.
	out = append(out, sortedKeys(want)...)
	return strings.Join(out, ",")
}

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func asList(v any) []any {
	l, _ := v.([]any)
	return l
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}

func optionalString(v any) *string {
	s, ok := v.(string)
	if !ok {
		return nil
	}
	return &s
}

func nullableString(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func quoteList(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(quoted, ", ")
}

// describeMismatch lists the mappings whose values read back differ from the ones written.
func describeMismatch(want, got map[string]objectMapping) []string {
	var out []string
	for _, target := range sortedKeys(want) {
		w := want[target]
		g, ok := got[target]
		if !ok {
			out = append(out, fmt.Sprintf("object mapping %q is missing", target))
			continue
		}
		for _, name := range sortedKeys(w.AttributeMappings) {
			wa := w.AttributeMappings[name]
			ga, ok := g.AttributeMappings[name]
			switch {
			case !ok:
				out = append(out, fmt.Sprintf("%s.%s is missing", target, name))
			case !reflect.DeepEqual(wa, ga):
				out = append(out, fmt.Sprintf("%s.%s: wrote %s, read %s", target, name, describeAttribute(wa), describeAttribute(ga)))
			}
		}
		for _, name := range sortedKeys(g.AttributeMappings) {
			if _, ok := w.AttributeMappings[name]; !ok {
				out = append(out, fmt.Sprintf("%s.%s was not removed", target, name))
			}
		}
		w.AttributeMappings, g.AttributeMappings = nil, nil
		if !reflect.DeepEqual(w, g) {
			out = append(out, fmt.Sprintf("object mapping %q settings differ (name, enabled, source object, flow types or scope)", target))
		}
	}
	for _, target := range sortedKeys(got) {
		if _, ok := want[target]; !ok {
			out = append(out, fmt.Sprintf("object mapping %q was not removed", target))
		}
	}
	return out
}

func describeAttribute(a attributeMapping) string {
	expr, def := "null", "null"
	if a.SourceExpression != nil {
		expr = fmt.Sprintf("%q", *a.SourceExpression)
	}
	if a.DefaultValue != nil {
		def = fmt.Sprintf("%q", *a.DefaultValue)
	}
	return fmt.Sprintf("{expression %s, default %s, flow %s/%s, matching %d, export missing references %t}",
		expr, def, a.FlowType, a.FlowBehavior, a.MatchingPriority, a.ExportMissingReferences)
}
