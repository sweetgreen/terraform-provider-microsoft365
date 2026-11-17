package graphBetaGroup

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/microsoft/kiota-abstractions-go/authentication"
	msgraph "github.com/microsoftgraph/msgraph-beta-sdk-go"
)

// fakeGroupsAPI serves GET /beta/groups from a fixed list of pages and records every request.
// A POST fails the test unless allowPost is set.
type fakeGroupsAPI struct {
	t      *testing.T
	server *httptest.Server
	pages  [][]string // each page holds JSON group objects
	status int        // non-zero: answer the list with this status and a Graph error
	// noContentType answers the list without a JSON content type, so Kiota returns no page.
	noContentType bool
	allowPost     bool

	mu      sync.Mutex
	filters []string
	posts   int
}

func newFakeGroupsAPI(t *testing.T, pages ...[]string) *fakeGroupsAPI {
	f := &fakeGroupsAPI{t: t, pages: pages}
	f.server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeGroupsAPI) serve(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	f.mu.Lock()
	defer f.mu.Unlock()

	if req.Method == http.MethodPost && req.URL.Path == "/beta/groups" {
		f.posts++
		if !f.allowPost {
			f.t.Errorf("unexpected POST /beta/groups: the group must not be created")
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":{"code":"Request_BadRequest","message":"not expected in this test"}}`)
		return
	}
	if req.Method != http.MethodGet || req.URL.Path != "/beta/groups" {
		f.t.Errorf("unexpected request: %s %s", req.Method, req.URL)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if f.noContentType {
		w.Header().Del("Content-Type")
		fmt.Fprint(w, `{"value":[]}`)
		return
	}
	if f.status != 0 {
		w.WriteHeader(f.status)
		fmt.Fprint(w, `{"error":{"code":"Authorization_RequestDenied","message":"Insufficient privileges to complete the operation."}}`)
		return
	}

	page := 0
	if p := req.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	} else {
		f.filters = append(f.filters, req.URL.Query().Get("$filter"))
	}
	var body strings.Builder
	body.WriteString("{")
	if page+1 < len(f.pages) {
		fmt.Fprintf(&body, `"@odata.nextLink":%q,`, fmt.Sprintf("%s/beta/groups?page=%d", f.server.URL, page+1))
	}
	values := []string{}
	if page < len(f.pages) {
		values = f.pages[page]
	}
	fmt.Fprintf(&body, `"value":[%s]}`, strings.Join(values, ","))
	fmt.Fprint(w, body.String())
}

func (f *fakeGroupsAPI) resource(t *testing.T) *GroupResource {
	t.Helper()
	adapter, err := msgraph.NewGraphRequestAdapter(&authentication.AnonymousAuthenticationProvider{})
	if err != nil {
		t.Fatal(err)
	}
	adapter.SetBaseUrl(f.server.URL + "/beta")
	r, ok := NewGroupResource().(*GroupResource)
	if !ok {
		t.Fatal("NewGroupResource did not return a *GroupResource")
	}
	r.client = msgraph.NewGraphServiceClient(adapter)
	return r
}

func groupJSON(id, displayName string) string {
	return fmt.Sprintf(`{"id":%q,"displayName":%q}`, id, displayName)
}

const (
	idA = "aaaaaaaa-0000-0000-0000-000000000001"
	idB = "bbbbbbbb-0000-0000-0000-000000000002"
	idC = "cccccccc-0000-0000-0000-000000000003"
)

func TestFindGroupsByDisplayName_ZeroOneAndManyMatches(t *testing.T) {
	cases := []struct {
		name  string
		pages [][]string
		want  []string
	}{
		{"zero matches", [][]string{{}}, nil},
		{"one match, different case", [][]string{{groupJSON(idA, "joe's lab devices")}}, []string{idA}},
		{
			"matches across pages, near-miss ignored",
			[][]string{
				{groupJSON(idA, "Joe's Lab Devices"), groupJSON(idC, "Joe's Lab Devices (old)")},
				{groupJSON(idB, "JOE'S LAB DEVICES")},
			},
			[]string{idA, idB},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			api := newFakeGroupsAPI(t, c.pages...)
			ids, err := api.resource(t).findGroupsByDisplayName(context.Background(), "Joe's Lab Devices")
			if err != nil {
				t.Fatalf("findGroupsByDisplayName: %v", err)
			}
			if fmt.Sprint(ids) != fmt.Sprint(c.want) {
				t.Errorf("ids = %v, want %v", ids, c.want)
			}
			if len(api.filters) != 1 || api.filters[0] != "displayName eq 'Joe''s Lab Devices'" {
				t.Errorf("Graph received $filter %q, want the escaped displayName filter", api.filters)
			}
		})
	}
}

func TestCheckDuplicateDisplayName_FailsClosedOnLookupError(t *testing.T) {
	api := newFakeGroupsAPI(t)
	api.status = http.StatusForbidden

	var diags diag.Diagnostics
	api.resource(t).checkDuplicateDisplayName(context.Background(), "POS Ring 1", &diags)

	if !diags.HasError() {
		t.Fatal("a failed lookup must stop the create")
	}
	if got := diags.Errors()[0].Summary(); got != duplicateLookupSummary {
		t.Errorf("summary = %q, want %q", got, duplicateLookupSummary)
	}
	if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, "Group.Read.All") {
		t.Errorf("detail %q should name the permission needed", detail)
	}
}

// guardedPlanModel is a new security group with prevent_duplicate_names set, as Create receives it.
func guardedPlanModel(displayName string) GroupResourceModel {
	return GroupResourceModel{
		ID:                            types.StringUnknown(),
		DisplayName:                   types.StringValue(displayName),
		Description:                   types.StringNull(),
		MailNickname:                  types.StringValue("posring1dup"),
		MailEnabled:                   types.BoolValue(false),
		SecurityEnabled:               types.BoolValue(true),
		GroupTypes:                    types.SetUnknown(types.StringType),
		Visibility:                    types.StringValue("Private"),
		IsAssignableToRole:            types.BoolValue(false),
		MembershipRule:                types.StringNull(),
		MembershipRuleProcessingState: types.StringUnknown(),
		CreatedDateTime:               types.StringUnknown(),
		GroupOwners:                   types.SetNull(types.StringType),
		GroupMembers:                  types.SetNull(types.StringType),
		HardDelete:                    types.BoolValue(false),
		PreventDuplicateNames:         types.BoolValue(true),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType, "read": types.StringType, "update": types.StringType, "delete": types.StringType,
		})},
	}
}

// Create checks again before it posts. ModifyPlan normally refuses first, but a name only known at
// apply time, or a group created by someone else after the plan, is caught here. This calls Create
// directly so ModifyPlan can't be the one refusing.
func TestCreate_RefusesDuplicateWithoutPosting(t *testing.T) {
	ctx := context.Background()
	api := newFakeGroupsAPI(t, []string{groupJSON(idA, "POS Ring 1 Windows Devices")})
	r := api.resource(t)

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	s := schemaResp.Schema
	nullValue := tftypes.NewValue(s.Type().TerraformType(ctx), nil)

	plan := tfsdk.Plan{Schema: s, Raw: nullValue}
	model := guardedPlanModel("pos ring 1 windows devices")
	if diags := plan.Set(ctx, &model); diags.HasError() {
		t.Fatalf("building the plan: %v", diags)
	}

	resp := resource.CreateResponse{State: tfsdk.State{Schema: s, Raw: nullValue}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("Create must refuse a group whose name is in use")
	}
	if got := resp.Diagnostics.Errors()[0].Summary(); got != duplicateNameSummary {
		t.Errorf("summary = %q, want %q", got, duplicateNameSummary)
	}
	if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, idA) {
		t.Errorf("detail %q should name the existing group %s", detail, idA)
	}
	if api.posts != 0 {
		t.Errorf("Create sent %d POST requests, want 0", api.posts)
	}
}

// Kiota returns (nil, nil) when it can't parse the list. That must fail the check, not read as
// "no duplicates" and not panic.
func TestFindGroupsByDisplayName_UnparseableResponseFailsClosed(t *testing.T) {
	api := newFakeGroupsAPI(t)
	api.noContentType = true

	var diags diag.Diagnostics
	api.resource(t).checkDuplicateDisplayName(context.Background(), "POS Ring 1", &diags)

	if !diags.HasError() || diags.Errors()[0].Summary() != duplicateLookupSummary {
		t.Fatalf("an unparseable list must fail the check, got %v", diags)
	}
}
