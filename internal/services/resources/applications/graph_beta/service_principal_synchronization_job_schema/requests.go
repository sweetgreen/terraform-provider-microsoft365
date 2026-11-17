package graphBetaServicePrincipalSynchronizationJobSchema

import (
	"context"
	"encoding/json"
	"fmt"

	abstractions "github.com/microsoft/kiota-abstractions-go"

	customrequests "github.com/deploymenttheory/terraform-provider-microsoft365/internal/services/common/custom_requests"
)

const schemaPath = "/servicePrincipals/{servicePrincipalId}/synchronization/jobs/{jobId}/schema"

func pathParameters(servicePrincipalID, jobID string) map[string]string {
	return map[string]string{"servicePrincipalId": servicePrincipalID, "jobId": jobID}
}

// getSchema returns the job's schema document as Graph returned it.
func (r *SynchronizationJobSchemaResource) getSchema(ctx context.Context, servicePrincipalID, jobID string) (map[string]any, error) {
	var raw json.RawMessage
	if err := customrequests.JSONRequest(
		ctx,
		r.client.GetAdapter(),
		abstractions.GET,
		schemaPath,
		pathParameters(servicePrincipalID, jobID),
		nil,
		&raw,
	); err != nil {
		return nil, fmt.Errorf("get synchronization schema: %w", err)
	}
	return decodeDocument(raw)
}

// putSchema replaces the job's schema. Graph answers 204 No Content.
func (r *SynchronizationJobSchemaResource) putSchema(ctx context.Context, servicePrincipalID, jobID string, body map[string]any) error {
	if err := customrequests.JSONRequest(
		ctx,
		r.client.GetAdapter(),
		abstractions.PUT,
		schemaPath,
		pathParameters(servicePrincipalID, jobID),
		body,
		nil,
	); err != nil {
		return fmt.Errorf("update synchronization schema: %w", err)
	}
	return nil
}

// getFunctions returns the expression functions the job's schema supports, by name.
func (r *SynchronizationJobSchemaResource) getFunctions(ctx context.Context, servicePrincipalID, jobID string) (map[string]functionSignature, error) {
	var response struct {
		Value []functionSignature `json:"value"`
	}
	if err := customrequests.JSONRequest(
		ctx,
		r.client.GetAdapter(),
		abstractions.GET,
		schemaPath+"/functions",
		pathParameters(servicePrincipalID, jobID),
		nil,
		&response,
	); err != nil {
		return nil, fmt.Errorf("get synchronization schema functions: %w", err)
	}
	out := make(map[string]functionSignature, len(response.Value))
	for _, f := range response.Value {
		out[f.ID] = f
	}
	return out, nil
}
