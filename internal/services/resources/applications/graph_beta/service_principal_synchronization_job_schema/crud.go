package graphBetaServicePrincipalSynchronizationJobSchema

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/constants"
	"github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/crud"
	errors "github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/errors/kiota"
)

func compositeID(m SynchronizationJobSchemaResourceModel) string {
	return strings.Join([]string{
		m.ServicePrincipalID.ValueString(),
		m.SynchronizationJobID.ValueString(),
		m.SynchronizationRuleID.ValueString(),
	}, "/")
}

func (r *SynchronizationJobSchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SynchronizationJobSchemaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := crud.HandleTimeout(ctx, plan.Timeouts.Create, CreateTimeout*time.Second, &resp.Diagnostics)
	if cancel == nil {
		return
	}
	defer cancel()

	desired, ok := r.write(ctx, plan, types.StringNull(), resp, constants.TfOperationCreate, &resp.Diagnostics)
	if !ok {
		return
	}
	plan.ID = types.StringValue(compositeID(plan))
	plan.Version = types.StringNull()
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	opts, lastRead := readbackOptions(plan, constants.TfOperationCreate)
	if err := crud.ReadWithRetry(
		ctx,
		r.Read,
		resource.ReadRequest{State: resp.State, ProviderMeta: req.ProviderMeta},
		&crud.CreateResponseContainer{CreateResponse: resp},
		opts,
	); err != nil {
		resp.Diagnostics.AddError("Error reading synchronization schema after create", readbackErrorDetail(err, desired, *lastRead))
	}
}

func (r *SynchronizationJobSchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SynchronizationJobSchemaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := crud.HandleTimeout(ctx, state.Timeouts.Read, ReadTimeout*time.Second, &resp.Diagnostics)
	if cancel == nil {
		return
	}
	defer cancel()

	doc, err := r.getSchema(ctx, state.ServicePrincipalID.ValueString(), state.SynchronizationJobID.ValueString())
	if err != nil {
		errors.HandleKiotaGraphErrorWithOptions(ctx, err, resp, constants.TfOperationRead, r.ReadPermissions,
			errors.GraphErrorOptions{PreserveStateOnReadBadRequest: true})
		return
	}

	if state.SynchronizationRuleID.IsNull() || state.SynchronizationRuleID.ValueString() == "" {
		// Imported without a rule ID: use the only rule there is.
		ids := ruleIDs(doc)
		if len(ids) != 1 {
			resp.Diagnostics.AddError(
				"Synchronization rule ID required",
				fmt.Sprintf("The schema has %d synchronization rules (%s). Import with service_principal_id/synchronization_job_id/synchronization_rule_id.",
					len(ids), quoteList(ids)),
			)
			return
		}
		state.SynchronizationRuleID = types.StringValue(ids[0])
	}

	snap, err := snapshot(doc, state.SynchronizationRuleID.ValueString())
	if err != nil {
		if len(ruleIDs(doc)) > 0 && !containsString(ruleIDs(doc), state.SynchronizationRuleID.ValueString()) {
			resp.Diagnostics.AddWarning(
				"Synchronization rule no longer exists",
				fmt.Sprintf("%s. Removing the resource from state; the next apply writes the configured mappings to the rule in configuration.", err),
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read synchronization schema", err.Error())
		return
	}

	mappings, diags := mappingsToTerraform(ctx, snap.Mappings, objectMappingModels(ctx, state.ObjectMappings))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ObjectMappings = mappings
	state.Version = types.StringPointerValue(snap.Version)
	state.ID = types.StringValue(compositeID(state))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SynchronizationJobSchemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SynchronizationJobSchemaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := crud.HandleTimeout(ctx, plan.Timeouts.Update, UpdateTimeout*time.Second, &resp.Diagnostics)
	if cancel == nil {
		return
	}
	defer cancel()

	desired, ok := r.write(ctx, plan, state.Version, resp, constants.TfOperationUpdate, &resp.Diagnostics)
	if !ok {
		return
	}
	plan.ID = types.StringValue(compositeID(plan))
	plan.Version = types.StringNull()
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	opts, lastRead := readbackOptions(plan, constants.TfOperationUpdate)
	if err := crud.ReadWithRetry(
		ctx,
		r.Read,
		resource.ReadRequest{State: resp.State, ProviderMeta: req.ProviderMeta},
		&crud.UpdateResponseContainer{UpdateResponse: resp},
		opts,
	); err != nil {
		resp.Diagnostics.AddError("Error reading synchronization schema after update", readbackErrorDetail(err, desired, *lastRead))
	}
}

// Delete leaves the mappings in Entra as they are. A provisioning job always has a
// schema, and resetting it to the application template's defaults could change what
// is provisioned (for example the ServiceNow `active` mapping), so destroy only stops
// Terraform from managing it.
func (r *SynchronizationJobSchemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Synchronization schema left unchanged",
		"The provisioning mappings were removed from Terraform state only. They are unchanged in Microsoft Entra.",
	)
	resp.State.RemoveResource(ctx)
}

// write reads the live schema, checks that it is still the version the plan was made
// from, replaces the rule's object mappings with the plan, and PUTs the document back.
func (r *SynchronizationJobSchemaResource) write(
	ctx context.Context,
	plan SynchronizationJobSchemaResourceModel,
	planVersion types.String,
	resp any,
	operation string,
	diags *diag.Diagnostics,
) (map[string]objectMapping, bool) {
	desired, d := mappingsFromTerraform(ctx, plan.ObjectMappings)
	diags.Append(d...)
	if diags.HasError() {
		return nil, false
	}
	spID, jobID := plan.ServicePrincipalID.ValueString(), plan.SynchronizationJobID.ValueString()

	doc, err := r.getSchema(ctx, spID, jobID)
	if err != nil {
		errors.HandleKiotaGraphError(ctx, err, resp, operation, r.ReadPermissions)
		return nil, false
	}
	if !planVersion.IsNull() && !planVersion.IsUnknown() {
		live, _ := doc["version"].(string)
		if live != planVersion.ValueString() {
			diags.AddError(
				"Synchronization schema changed since the plan",
				fmt.Sprintf("The plan was made from schema version %q, but Entra now has version %q, so someone changed "+
					"the mappings after the plan. Nothing was written. Run plan again and review the difference before applying.",
					planVersion.ValueString(), live),
			)
			return nil, false
		}
	}

	var functionErr error
	body, err := buildDocument(doc, plan.SynchronizationRuleID.ValueString(), desired, func() (map[string]functionSignature, error) {
		sigs, err := r.getFunctions(ctx, spID, jobID)
		functionErr = err
		return sigs, err
	})
	if functionErr != nil {
		errors.HandleKiotaGraphError(ctx, functionErr, resp, operation, r.ReadPermissions)
		return nil, false
	}
	if err != nil {
		diags.AddError("Invalid synchronization schema configuration", "Nothing was written: "+err.Error())
		return nil, false
	}

	if err := r.putSchema(ctx, spID, jobID, body); err != nil {
		errors.HandleKiotaGraphError(ctx, err, resp, operation, r.WritePermissions)
		return nil, false
	}
	return desired, true
}

// readbackOptions waits until a read returns the mappings that were written; Entra
// can serve a stale schema straight after a PUT. The returned pointer holds the last
// mappings read, for the error message.
func readbackOptions(plan SynchronizationJobSchemaResourceModel, operation string) (crud.ReadWithRetryOptions, *map[string]objectMapping) {
	lastRead := map[string]objectMapping{}
	opts := crud.DefaultReadWithRetryOptions()
	opts.Operation = operation
	opts.ResourceTypeName = ResourceName
	opts.ConsistencyPredicate = func(ctx context.Context, state tfsdk.State) bool {
		var actual SynchronizationJobSchemaResourceModel
		if state.Get(ctx, &actual).HasError() {
			return false
		}
		if read, d := mappingsFromTerraform(ctx, actual.ObjectMappings); !d.HasError() {
			lastRead = read
		}
		return actual.ObjectMappings.Equal(plan.ObjectMappings)
	}
	return opts, &lastRead
}

func readbackErrorDetail(err error, desired, lastRead map[string]objectMapping) string {
	detail := "Entra accepted the new schema, but reading it back did not return the configured mappings: " + err.Error()
	if diffs := describeMismatch(desired, lastRead); len(diffs) > 0 {
		detail += "\n\nDifferences in the last read:\n  - " + strings.Join(diffs, "\n  - ")
	}
	return detail
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
