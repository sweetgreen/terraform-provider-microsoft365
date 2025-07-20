package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/jarcoal/httpmock"
	"github.com/sweetgreen/terraform-provider-microsoft365/internal/mocks/factories"
)

// mockState tracks the state of resources for consistent responses
var mockState struct {
	sync.Mutex
	scripts     map[string]map[string]interface{}
	assignments map[string][]map[string]interface{}
}

func init() {
	// Initialize mockState
	mockState.scripts = make(map[string]map[string]interface{})
	mockState.assignments = make(map[string][]map[string]interface{})

	// Register a default 404 responder for any unmatched requests
	httpmock.RegisterNoResponder(httpmock.NewStringResponder(404, `{"error":{"code":"ResourceNotFound","message":"Resource not found"}}`))
}

// MacosCustomAttributeScriptAssignmentMock provides mock responses for macOS custom attribute script assignment operations
type MacosCustomAttributeScriptAssignmentMock struct{}

// RegisterMocks registers HTTP mock responses for macOS custom attribute script assignment operations
func (m *MacosCustomAttributeScriptAssignmentMock) RegisterMocks() {
	// Reset the state when registering mocks
	mockState.Lock()
	mockState.scripts = make(map[string]map[string]interface{})
	mockState.assignments = make(map[string][]map[string]interface{})
	mockState.Unlock()

	// Initialize base script data
	baseScriptId := "00000000-0000-0000-0000-000000000002"
	baseScriptData := map[string]interface{}{
		"id":                   baseScriptId,
		"displayName":          "Test macOS Custom Attribute Script",
		"description":          "Test description",
		"scriptContent":        "#!/bin/bash\necho \"test\"",
		"createdDateTime":      "2024-01-01T00:00:00Z",
		"lastModifiedDateTime": "2024-01-01T00:00:00Z",
		"runAsAccount":         "system",
		"fileName":             "test.sh",
	}

	mockState.Lock()
	mockState.scripts[baseScriptId] = baseScriptData
	mockState.assignments[baseScriptId] = []map[string]interface{}{}
	mockState.Unlock()

	// Register GET for script data
	httpmock.RegisterResponder("GET", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceCustomAttributeShellScripts/[^/]+$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			scriptId := urlParts[len(urlParts)-1]

			mockState.Lock()
			scriptData, exists := mockState.scripts[scriptId]
			mockState.Unlock()

			if !exists {
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Script not found"}}`), nil
			}

			return httpmock.NewJsonResponse(200, scriptData)
		})

	// Register GET for script assignments (list)
	httpmock.RegisterResponder("GET", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceCustomAttributeShellScripts/[^/]+/assignments$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			scriptId := urlParts[len(urlParts)-2]

			mockState.Lock()
			_, exists := mockState.scripts[scriptId]
			assignments := mockState.assignments[scriptId]
			mockState.Unlock()

			if !exists {
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Script not found"}}`), nil
			}

			response := map[string]interface{}{
				"@odata.context": fmt.Sprintf("https://graph.microsoft.com/beta/$metadata#deviceManagement/deviceCustomAttributeShellScripts('%s')/assignments", scriptId),
				"value":          assignments,
			}

			return httpmock.NewJsonResponse(200, response)
		})

	// Register GET for specific assignment
	httpmock.RegisterResponder("GET", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceCustomAttributeShellScripts/[^/]+/assignments/[^/]+$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			assignmentId := urlParts[len(urlParts)-1]
			scriptId := urlParts[len(urlParts)-3]

			mockState.Lock()
			assignments := mockState.assignments[scriptId]
			mockState.Unlock()

			for _, assignment := range assignments {
				if assignment["id"] == assignmentId {
					return httpmock.NewJsonResponse(200, assignment)
				}
			}

			return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Assignment not found"}}`), nil
		})

	// Register POST for creating assignment
	httpmock.RegisterResponder("POST", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceCustomAttributeShellScripts/[^/]+/assignments$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			scriptId := urlParts[len(urlParts)-2]

			mockState.Lock()
			_, exists := mockState.scripts[scriptId]
			mockState.Unlock()

			if !exists {
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Script not found"}}`), nil
			}

			// Parse request body
			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				return httpmock.NewStringResponse(400, `{"error":{"code":"BadRequest","message":"Invalid request body"}}`), nil
			}

			// Create new assignment
			newAssignment := map[string]interface{}{
				"id":     uuid.New().String(),
				"target": requestBody["target"],
			}

			mockState.Lock()
			mockState.assignments[scriptId] = append(mockState.assignments[scriptId], newAssignment)
			mockState.Unlock()

			return httpmock.NewJsonResponse(201, newAssignment)
		})

	// Register PATCH for updating assignment
	httpmock.RegisterResponder("PATCH", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceCustomAttributeShellScripts/[^/]+/assignments/[^/]+$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			assignmentId := urlParts[len(urlParts)-1]
			scriptId := urlParts[len(urlParts)-3]

			mockState.Lock()
			assignments := mockState.assignments[scriptId]
			mockState.Unlock()

			// Parse request body
			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				return httpmock.NewStringResponse(400, `{"error":{"code":"BadRequest","message":"Invalid request body"}}`), nil
			}

			// Update assignment
			for i, assignment := range assignments {
				if assignment["id"] == assignmentId {
					// Update the assignment
					if target, ok := requestBody["target"]; ok {
						assignment["target"] = target
					}

					mockState.Lock()
					mockState.assignments[scriptId][i] = assignment
					mockState.Unlock()

					return httpmock.NewJsonResponse(200, assignment)
				}
			}

			return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Assignment not found"}}`), nil
		})

	// Register DELETE for assignment
	httpmock.RegisterResponder("DELETE", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceCustomAttributeShellScripts/[^/]+/assignments/[^/]+$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			assignmentId := urlParts[len(urlParts)-1]
			scriptId := urlParts[len(urlParts)-3]

			mockState.Lock()
			assignments := mockState.assignments[scriptId]

			// Find and remove the assignment
			newAssignments := []map[string]interface{}{}
			found := false
			for _, assignment := range assignments {
				if assignment["id"] != assignmentId {
					newAssignments = append(newAssignments, assignment)
				} else {
					found = true
				}
			}

			if found {
				mockState.assignments[scriptId] = newAssignments
			}
			mockState.Unlock()

			if !found {
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Assignment not found"}}`), nil
			}

			return httpmock.NewStringResponse(204, ""), nil
		})

	// Register POST for assign action (alternative endpoint)
	httpmock.RegisterResponder("POST", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceCustomAttributeShellScripts/[^/]+/assign$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			scriptId := urlParts[len(urlParts)-2]

			mockState.Lock()
			_, exists := mockState.scripts[scriptId]
			mockState.Unlock()

			if !exists {
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Script not found"}}`), nil
			}

			// Parse request body
			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				return httpmock.NewStringResponse(400, `{"error":{"code":"BadRequest","message":"Invalid request body"}}`), nil
			}

			// Process assignments from the request
			if assignments, ok := requestBody["assignments"].([]interface{}); ok {
				mockState.Lock()
				// Clear existing assignments
				mockState.assignments[scriptId] = []map[string]interface{}{}

				// Add new assignments
				for _, assignment := range assignments {
					if assignmentObj, ok := assignment.(map[string]interface{}); ok {
						// Ensure assignment has an ID
						if assignmentObj["id"] == nil || assignmentObj["id"] == "" {
							assignmentObj["id"] = uuid.New().String()
						}
						mockState.assignments[scriptId] = append(mockState.assignments[scriptId], assignmentObj)
					}
				}
				mockState.Unlock()
			}

			return httpmock.NewStringResponse(204, ""), nil
		})

	// Register additional scripts for testing
	registerSpecificScriptMocks()
}

// RegisterErrorMocks registers HTTP mock responses for error scenarios
func (m *MacosCustomAttributeScriptAssignmentMock) RegisterErrorMocks() {
	// Register error response for script assignment
	errorScriptId := "99999999-9999-9999-9999-999999999999"
	httpmock.RegisterResponder("POST", "https://graph.microsoft.com/beta/deviceManagement/deviceCustomAttributeShellScripts/"+errorScriptId+"/assignments",
		factories.ErrorResponse(400, "BadRequest", "Error creating assignment"))

	// Register GET for error script to ensure it exists but will fail on assignment
	httpmock.RegisterResponder("GET", "https://graph.microsoft.com/beta/deviceManagement/deviceCustomAttributeShellScripts/"+errorScriptId,
		func(req *http.Request) (*http.Response, error) {
			scriptData := map[string]interface{}{
				"id":          errorScriptId,
				"displayName": "Error Test Script",
				"description": "Script that fails on assignment",
			}
			return httpmock.NewJsonResponse(200, scriptData)
		})

	// Register error response for script not found
	httpmock.RegisterResponder("GET", "https://graph.microsoft.com/beta/deviceManagement/deviceCustomAttributeShellScripts/not-found-script",
		factories.ErrorResponse(404, "ResourceNotFound", "Script not found"))
}

// registerSpecificScriptMocks registers mocks for specific test scenarios
func registerSpecificScriptMocks() {
	// Minimal script with no assignments
	minimalScriptId := "00000000-0000-0000-0000-000000000003"
	minimalScriptData := map[string]interface{}{
		"id":                   minimalScriptId,
		"displayName":          "Minimal Test Script",
		"description":          "Minimal test description",
		"scriptContent":        "#!/bin/bash\necho \"minimal\"",
		"createdDateTime":      "2024-01-01T00:00:00Z",
		"lastModifiedDateTime": "2024-01-01T00:00:00Z",
		"runAsAccount":         "system",
		"fileName":             "minimal.sh",
	}

	mockState.Lock()
	mockState.scripts[minimalScriptId] = minimalScriptData
	mockState.assignments[minimalScriptId] = []map[string]interface{}{}
	mockState.Unlock()

	// Maximal script with multiple assignments
	maximalScriptId := "00000000-0000-0000-0000-000000000004"
	maximalScriptData := map[string]interface{}{
		"id":                   maximalScriptId,
		"displayName":          "Maximal Test Script",
		"description":          "Maximal test description with all features",
		"scriptContent":        "#!/bin/bash\necho \"maximal\"",
		"createdDateTime":      "2024-01-01T00:00:00Z",
		"lastModifiedDateTime": "2024-01-01T00:00:00Z",
		"runAsAccount":         "system",
		"fileName":             "maximal.sh",
	}

	maximalAssignments := []map[string]interface{}{
		{
			"id": "11111111-1111-1111-1111-111111111111",
			"target": map[string]interface{}{
				"@odata.type": "#microsoft.graph.groupAssignmentTarget",
				"groupId":     "22222222-2222-2222-2222-222222222222",
				"deviceAndAppManagementAssignmentFilterId":   "33333333-3333-3333-3333-333333333333",
				"deviceAndAppManagementAssignmentFilterType": "include",
			},
		},
		{
			"id": "44444444-4444-4444-4444-444444444444",
			"target": map[string]interface{}{
				"@odata.type": "#microsoft.graph.allDevicesAssignmentTarget",
			},
		},
	}

	mockState.Lock()
	mockState.scripts[maximalScriptId] = maximalScriptData
	mockState.assignments[maximalScriptId] = maximalAssignments
	mockState.Unlock()
}
