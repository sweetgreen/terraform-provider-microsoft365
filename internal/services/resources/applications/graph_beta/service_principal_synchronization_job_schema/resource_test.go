package graphBetaServicePrincipalSynchronizationJobSchema_test

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"

	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/acceptance/check"
	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/helpers"
	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/mocks"
	target "github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/resources/applications/graph_beta/service_principal_synchronization_job_schema"
	testmocks "github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/resources/applications/graph_beta/service_principal_synchronization_job_schema/mocks"
)

const resourceAddress = target.ResourceName + ".test"

func setup(t *testing.T) *testmocks.Mock {
	t.Helper()
	mocks.SetupUnitTestEnvironment(t)
	httpmock.Activate()
	mocks.NewMocks().AuthMocks.RegisterMocks()
	m := &testmocks.Mock{}
	m.RegisterMocks()
	t.Cleanup(func() { httpmock.DeactivateAndReset(); m.CleanupMockState() })
	return m
}

func config(t *testing.T, name string) string {
	t.Helper()
	cfg, err := helpers.ParseHCLFile("tests/terraform/unit/" + name)
	require.NoError(t, err)
	return cfg
}

func objectMappings(t *testing.T, doc map[string]any) map[string]map[string]any {
	t.Helper()
	rules, _ := doc["synchronizationRules"].([]any)
	require.Len(t, rules, 1)
	out := map[string]map[string]any{}
	for _, om := range rules[0].(map[string]any)["objectMappings"].([]any) {
		m := om.(map[string]any)
		attrs := map[string]any{}
		for _, am := range m["attributeMappings"].([]any) {
			a := am.(map[string]any)
			attrs[a["targetAttributeName"].(string)] = a
		}
		m["byName"] = attrs
		out[m["targetObjectName"].(string)] = m
	}
	return out
}

// Adopting the live schema, importing it and changing one mapping.
func TestUnitResourceSynchronizationJobSchema_01_Lifecycle(t *testing.T) {
	m := setup(t)
	original := m.Schema
	initial := config(t, "resource.tf")
	updated := config(t, "resource_updated.tf")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: initial,
				Check: resource.ComposeTestCheckFunc(
					check.That(resourceAddress).Key("id").HasValue(testmocks.ServicePrincipalID+"/"+testmocks.JobID+"/"+testmocks.RuleID),
					check.That(resourceAddress).Key("object_mappings.sys_user.attribute_mappings.active.source_expression").
						HasValue(`Switch([IsSoftDeleted], , "False", "1", "True", "0")`),
					check.That(resourceAddress).Key("object_mappings.sys_user.attribute_mappings.email.flow_type").HasValue("Always"),
					check.That(resourceAddress).Key("object_mappings.cmn_cost_center.name").DoesNotExist(),
					check.That(resourceAddress).Key("version").HasValue("Date:2026-11-17T00:00:00.0000000Z, ActivityId:put-1"),
				),
			},
			{
				// The rule ID can be left out when the schema has one rule.
				ResourceName:            resourceAddress,
				ImportState:             true,
				ImportStateId:           testmocks.ServicePrincipalID + "/" + testmocks.JobID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
			{Config: initial, PlanOnly: true},
			{
				Config: updated,
				Check: check.That(resourceAddress).Key("object_mappings.sys_user.attribute_mappings.active.source_expression").
					HasValue(`Switch([IsSoftDeleted], "1", "True", "0")`),
			},
			{Config: updated, PlanOnly: true},
		},
	})

	m.Lock()
	defer m.Unlock()
	require.Len(t, m.Puts, 2)
	require.Equal(t, 1, m.FunctionGets, "function definitions are only needed for the changed function expressions")

	// Adopting an identical configuration sends Graph its own document back.
	first, _ := json.Marshal(m.Puts[0])
	want := map[string]any{}
	for k, v := range original {
		if k != "@odata.context" && k != "directories@odata.context" {
			want[k] = v
		}
	}
	wantJSON, _ := json.Marshal(want)
	require.JSONEq(t, string(wantJSON), string(first))

	put := objectMappings(t, m.Puts[1])
	require.NotContains(t, put, "cmn_cost_center")
	require.Equal(t, original["directories"], m.Puts[1]["directories"])
	user := put["sys_user"]["byName"].(map[string]any)
	require.NotContains(t, user, "location")

	active := user["active"].(map[string]any)
	require.Equal(t, "0", active["defaultValue"])
	src := active["source"].(map[string]any)
	require.Equal(t, `Switch([IsSoftDeleted], "1", "True", "0")`, src["expression"])
	require.Equal(t, "Function", src["type"])
	var keys []string
	for _, p := range src["parameters"].([]any) {
		keys = append(keys, p.(map[string]any)["key"].(string))
	}
	require.Equal(t, []string{"source", "defaultValue", "switchValue", "switchValue"}, keys)

	title := user["title"].(map[string]any)
	require.Equal(t, "ObjectAddOnly", title["flowType"])
	require.Equal(t, map[string]any{"expression": "[jobTitle]", "name": "jobTitle", "parameters": []any{}, "type": "Attribute"}, title["source"])

	// Destroy leaves the schema in Entra alone.
	require.Equal(t, m.Puts[1]["synchronizationRules"], m.Schema["synchronizationRules"])
}

// An expression Entra would store differently is rejected at plan time.
func TestUnitResourceSynchronizationJobSchema_02_NonCanonicalExpression(t *testing.T) {
	m := setup(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config(t, "resource_noncanonical.tf"),
				ExpectError: regexp.MustCompile(`(?s)Entra stores this expression as.*Coalesce\(\[mail\], \[userPrincipalName\]\)`),
			},
		},
	})
	require.Empty(t, m.Puts)
}

// A change made in the portal after the plan stops the apply before anything is written.
func TestUnitResourceSynchronizationJobSchema_03_SchemaChangedSincePlan(t *testing.T) {
	m := setup(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config(t, "resource.tf")},
			{
				PreConfig:   m.ChangeVersionOnEveryGet,
				Config:      config(t, "resource_updated.tf"),
				ExpectError: regexp.MustCompile(`Synchronization schema changed since the plan`),
			},
		},
	})
	m.Lock()
	defer m.Unlock()
	require.Len(t, m.Puts, 1, "only the first step may write")
}

// Entra can return the previous schema straight after a PUT; the provider keeps reading.
func TestUnitResourceSynchronizationJobSchema_04_StaleReadAfterPut(t *testing.T) {
	m := setup(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: mocks.TestUnitTestProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config(t, "resource.tf")},
			{
				PreConfig: func() { m.SetStaleReadsAfterPut(1) },
				Config:    config(t, "resource_updated.tf"),
				Check: check.That(resourceAddress).Key("object_mappings.sys_user.attribute_mappings.active.default_value").
					HasValue("0"),
			},
			{Config: config(t, "resource_updated.tf"), PlanOnly: true},
		},
	})
}
