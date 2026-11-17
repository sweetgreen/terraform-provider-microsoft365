package graphBetaGroup

import (
	"context"
	"fmt"
	"strings"

	errors "github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/errors/kiota"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/groups"
)

// Microsoft Graph lets two groups share a display name, although the Entra admin UI does not. With
// prevent_duplicate_names set, the resource looks the name up before it creates a group, at plan
// time when the name is known and again in Create, and refuses if any group already has it. It
// never looks on update, refresh or import, so an existing group can always be adopted.

const (
	duplicateNameSummary   = "Group display name already in use"
	duplicateLookupSummary = "Unable to check for groups with the same display name"
)

// escapeODataString escapes a value for use inside a single-quoted OData string literal.
func escapeODataString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

// displayNameFilter returns the $filter that selects groups with the given display name.
func displayNameFilter(displayName string) string {
	return fmt.Sprintf("displayName eq '%s'", escapeODataString(displayName))
}

// findGroupsByDisplayName returns the object IDs of every group whose display name equals
// displayName, ignoring case as Entra does. It follows @odata.nextLink so every match is reported.
func (r *GroupResource) findGroupsByDisplayName(ctx context.Context, displayName string) ([]string, error) {
	filter := displayNameFilter(displayName)
	requestConfig := &groups.GroupsRequestBuilderGetRequestConfiguration{
		QueryParameters: &groups.GroupsRequestBuilderGetQueryParameters{
			Filter: &filter,
			Select: []string{"id", "displayName"},
		},
	}

	tflog.Debug(ctx, "Looking for groups with the same display name", map[string]any{"filter": filter})

	builder := r.client.Groups()
	page, err := builder.Get(ctx, requestConfig)
	var ids []string
	for {
		if err != nil {
			return nil, err
		}
		// Kiota returns no page and no error when the body can't be parsed, for example without a
		// JSON content type. Treat that as a failed lookup rather than as "no matches".
		if page == nil {
			return nil, fmt.Errorf("no parseable group list returned by Graph for %s", filter)
		}
		for _, group := range page.GetValue() {
			if group.GetId() == nil || group.GetDisplayName() == nil {
				continue
			}
			if strings.EqualFold(*group.GetDisplayName(), displayName) {
				ids = append(ids, *group.GetId())
			}
		}
		next := page.GetOdataNextLink()
		if next == nil || *next == "" {
			return ids, nil
		}
		page, err = builder.WithUrl(*next).Get(ctx, nil)
	}
}

// duplicateNameDetail explains which groups already hold the name and how to adopt one.
func duplicateNameDetail(displayName string, ids []string) string {
	importHint := "an import block (import { to = <resource address>, id = \"<object id>\" }) or terraform import"
	return fmt.Sprintf(
		"prevent_duplicate_names is true and the tenant already has %d group(s) named %q: %s. "+
			"Adopt the existing group with %s, rename this group, or set prevent_duplicate_names = false "+
			"to create another group with the same name.",
		len(ids), displayName, strings.Join(ids, ", "), importHint,
	)
}

// checkDuplicateDisplayName adds an error to diags when a group named displayName already exists
// or the lookup fails. It fails closed: a lookup that can't be made never lets the create through.
func (r *GroupResource) checkDuplicateDisplayName(ctx context.Context, displayName string, diags *diag.Diagnostics) {
	ids, err := r.findGroupsByDisplayName(ctx, displayName)
	if err != nil {
		info := errors.GraphError(ctx, err)
		detail := err.Error()
		if info.ErrorMessage != "" {
			detail = fmt.Sprintf("Graph returned %d %s: %s", info.StatusCode, info.ErrorCode, info.ErrorMessage)
		}
		diags.AddAttributeError(
			path.Root("display_name"),
			duplicateLookupSummary,
			fmt.Sprintf("prevent_duplicate_names is true, so the provider must list groups named %q before creating one, "+
				"and the lookup failed. Listing groups needs Group.Read.All or Directory.Read.All. %s",
				displayName, detail),
		)
		return
	}

	if len(ids) > 0 {
		diags.AddAttributeError(path.Root("display_name"), duplicateNameSummary, duplicateNameDetail(displayName, ids))
	}
}
