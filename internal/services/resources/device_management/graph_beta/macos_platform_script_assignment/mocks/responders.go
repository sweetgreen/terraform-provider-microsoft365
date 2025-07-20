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

// MacosPlatformScriptAssignmentMock provides mock responses for macOS platform script assignment operations
type MacosPlatformScriptAssignmentMock struct{}

// RegisterMocks registers HTTP mock responses for macOS platform script assignment operations
func (m *MacosPlatformScriptAssignmentMock) RegisterMocks() {
	// Reset the state when registering mocks
	mockState.Lock()
	mockState.scripts = make(map[string]map[string]interface{})
	mockState.assignments = make(map[string][]map[string]interface{})
	mockState.Unlock()

	// Initialize base script data
	baseScriptId := "00000000-0000-0000-0000-000000000002"
	baseScriptData := map[string]interface{}{
		"id":                          baseScriptId,
		"displayName":                 "Test macOS Script",
		"description":                 "A test macOS platform script",
		"fileName":                    "test-script.sh",
		"scriptContent":               "#!/bin/bash\necho 'Hello, World!'",
		"runAsAccount":                "system",
		"blockExecutionNotifications": false,
		"enforceSignatureCheck":       false,
	}

	mockState.Lock()
	mockState.scripts[baseScriptId] = baseScriptData
	mockState.assignments[baseScriptId] = []map[string]interface{}{}
	mockState.Unlock()

	// Register GET for script
	httpmock.RegisterResponder("GET", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/[^/]+$`,
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

	// Register GET for script assignments
	httpmock.RegisterResponder("GET", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/[^/]+/assignments$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			scriptId := urlParts[len(urlParts)-2]

			mockState.Lock()
			assignments, exists := mockState.assignments[scriptId]
			mockState.Unlock()

			if !exists {
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Script not found"}}`), nil
			}

			response := map[string]interface{}{
				"@odata.context": fmt.Sprintf("https://graph.microsoft.com/beta/$metadata#deviceManagement/deviceShellScripts('%s')/assignments", scriptId),
				"value":          assignments,
			}

			return httpmock.NewJsonResponse(200, response)
		})

	// Register GET for specific assignment
	httpmock.RegisterResponder("GET", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/[^/]+/assignments/[^/]+$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			assignmentId := urlParts[len(urlParts)-1]
			scriptId := urlParts[len(urlParts)-3]

			mockState.Lock()
			assignments, exists := mockState.assignments[scriptId]
			mockState.Unlock()

			if !exists {
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Script not found"}}`), nil
			}

			for _, assignment := range assignments {
				if assignment["id"] == assignmentId {
					// Ensure the assignment has a target
					if assignment["target"] == nil {
						assignment["target"] = map[string]interface{}{
							"@odata.type": "#microsoft.graph.allDevicesAssignmentTarget",
						}
					}
					return httpmock.NewJsonResponse(200, assignment)
				}
			}

			return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Assignment not found"}}`), nil
		})

	// Register PATCH for specific assignment update
	httpmock.RegisterResponder("PATCH", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/[^/]+/assignments/[^/]+$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			assignmentId := urlParts[len(urlParts)-1]
			scriptId := urlParts[len(urlParts)-3]

			mockState.Lock()
			assignments, exists := mockState.assignments[scriptId]
			if !exists {
				mockState.Unlock()
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Script not found"}}`), nil
			}

			// Parse request body
			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				mockState.Unlock()
				return httpmock.NewStringResponse(400, `{"error":{"code":"BadRequest","message":"Invalid request body"}}`), nil
			}

			// Find and update the assignment
			found := false
			for i, assignment := range assignments {
				if assignment["id"] == assignmentId {
					// Update the target
					if target, ok := requestBody["target"]; ok {
						assignments[i]["target"] = target
					}
					found = true
					mockState.assignments[scriptId] = assignments
					mockState.Unlock()
					return httpmock.NewJsonResponse(200, assignments[i])
				}
			}
			mockState.Unlock()

			if !found {
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Assignment not found"}}`), nil
			}

			return nil, nil
		})

	// Register POST for script assignment
	httpmock.RegisterResponder("POST", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/[^/]+/assignments$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			scriptId := urlParts[len(urlParts)-2]

			mockState.Lock()
			_, exists := mockState.scripts[scriptId]
			if !exists {
				mockState.Unlock()
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Script not found"}}`), nil
			}
			mockState.Unlock()

			// Parse request body
			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				return httpmock.NewStringResponse(400, `{"error":{"code":"BadRequest","message":"Invalid request body"}}`), nil
			}

			// Create new assignment
			assignmentId := uuid.New().String()
			newAssignment := map[string]interface{}{
				"id":     assignmentId,
				"target": requestBody["target"],
			}

			// Add to assignments
			mockState.Lock()
			if mockState.assignments[scriptId] == nil {
				mockState.assignments[scriptId] = []map[string]interface{}{}
			}
			mockState.assignments[scriptId] = append(mockState.assignments[scriptId], newAssignment)
			mockState.Unlock()

			return httpmock.NewJsonResponse(201, newAssignment)
		})

	// Register POST for batch assignment using /assign endpoint
	httpmock.RegisterResponder("POST", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/[^/]+/assign$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			scriptId := urlParts[len(urlParts)-2]

			mockState.Lock()
			_, exists := mockState.scripts[scriptId]
			if !exists {
				mockState.Unlock()
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Script not found"}}`), nil
			}
			mockState.Unlock()

			// Parse request body
			var requestBody map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&requestBody)
			if err != nil {
				return httpmock.NewStringResponse(400, `{"error":{"code":"BadRequest","message":"Invalid request body"}}`), nil
			}

			// Process assignments
			if assignments, ok := requestBody["deviceManagementScriptAssignments"].([]interface{}); ok {
				mockState.Lock()
				// Replace all assignments
				mockState.assignments[scriptId] = []map[string]interface{}{}

				for _, assignment := range assignments {
					if assignmentObj, ok := assignment.(map[string]interface{}); ok {
						newAssignment := map[string]interface{}{
							"id":     uuid.New().String(),
							"target": assignmentObj["target"],
						}
						mockState.assignments[scriptId] = append(mockState.assignments[scriptId], newAssignment)
					}
				}
				mockState.Unlock()
			}

			// Return collection of assignments
			mockState.Lock()
			assignments := mockState.assignments[scriptId]
			mockState.Unlock()

			response := map[string]interface{}{
				"value": assignments,
			}

			return httpmock.NewJsonResponse(200, response)
		})

	// Register DELETE for specific assignment
	httpmock.RegisterResponder("DELETE", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/[^/]+/assignments/[^/]+$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			assignmentId := urlParts[len(urlParts)-1]
			scriptId := urlParts[len(urlParts)-3]

			mockState.Lock()
			assignments, exists := mockState.assignments[scriptId]
			if !exists {
				mockState.Unlock()
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Script not found"}}`), nil
			}

			// Find and remove the assignment
			found := false
			newAssignments := []map[string]interface{}{}
			for _, assignment := range assignments {
				if assignment["id"] == assignmentId {
					found = true
				} else {
					newAssignments = append(newAssignments, assignment)
				}
			}

			if !found {
				mockState.Unlock()
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Assignment not found"}}`), nil
			}

			mockState.assignments[scriptId] = newAssignments
			mockState.Unlock()

			return httpmock.NewStringResponse(204, ""), nil
		})

	// Register DELETE for specific assignment using groupAssignments endpoint
	httpmock.RegisterResponder("DELETE", `=~^https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/[^/]+/groupAssignments/[^/]+$`,
		func(req *http.Request) (*http.Response, error) {
			urlParts := strings.Split(req.URL.Path, "/")
			assignmentId := urlParts[len(urlParts)-1]
			scriptId := urlParts[len(urlParts)-3]

			mockState.Lock()
			assignments, exists := mockState.assignments[scriptId]
			if !exists {
				mockState.Unlock()
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Script not found"}}`), nil
			}

			// Find and remove the assignment
			found := false
			newAssignments := []map[string]interface{}{}
			for _, assignment := range assignments {
				if assignment["id"] == assignmentId {
					found = true
				} else {
					newAssignments = append(newAssignments, assignment)
				}
			}

			if !found {
				mockState.Unlock()
				return httpmock.NewStringResponse(404, `{"error":{"code":"ResourceNotFound","message":"Assignment not found"}}`), nil
			}

			mockState.assignments[scriptId] = newAssignments
			mockState.Unlock()

			return httpmock.NewStringResponse(204, ""), nil
		})

	// Register specific scripts for testing
	registerSpecificScriptMocks()
}

// RegisterErrorMocks registers HTTP mock responses for error scenarios
func (m *MacosPlatformScriptAssignmentMock) RegisterErrorMocks() {
	// Register error response for script assignment
	errorScriptId := "99999999-9999-9999-9999-999999999999"
	httpmock.RegisterResponder("POST", "https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/"+errorScriptId+"/assignments",
		factories.ErrorResponse(400, "BadRequest", "Error creating assignment"))

	// Register GET for error script to ensure it exists but will fail on assignment
	httpmock.RegisterResponder("GET", "https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/"+errorScriptId,
		func(req *http.Request) (*http.Response, error) {
			scriptData := map[string]interface{}{
				"id":          errorScriptId,
				"displayName": "Error Test Script",
				"description": "A script for testing errors",
			}
			return httpmock.NewJsonResponse(200, scriptData)
		})

	// Register error response for script not found
	notFoundScriptId := "ffffffff-ffff-ffff-ffff-ffffffffffff"
	httpmock.RegisterResponder("GET", "https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/"+notFoundScriptId,
		factories.ErrorResponse(404, "ResourceNotFound", "Script not found"))

	// Also register the assignments endpoint for not found script
	httpmock.RegisterResponder("POST", "https://graph.microsoft.com/beta/deviceManagement/deviceShellScripts/"+notFoundScriptId+"/assignments",
		factories.ErrorResponse(404, "ResourceNotFound", "Script not found"))
}

// registerSpecificScriptMocks registers mocks for specific test scenarios
func registerSpecificScriptMocks() {
	// Minimal script with no assignments
	minimalScriptId := "00000000-0000-0000-0000-000000000003"
	minimalScriptData := map[string]interface{}{
		"id":            minimalScriptId,
		"displayName":   "Minimal Test Script",
		"description":   "A minimal test script",
		"fileName":      "minimal.sh",
		"scriptContent": "#!/bin/bash\necho 'Minimal'",
		"runAsAccount":  "system",
	}

	mockState.Lock()
	mockState.scripts[minimalScriptId] = minimalScriptData
	mockState.assignments[minimalScriptId] = []map[string]interface{}{}
	mockState.Unlock()

	// Maximal script with multiple assignments
	maximalScriptId := "00000000-0000-0000-0000-000000000004"
	maximalScriptData := map[string]interface{}{
		"id":                          maximalScriptId,
		"displayName":                 "Maximal Test Script",
		"description":                 "A maximal test script with all options",
		"fileName":                    "maximal.sh",
		"scriptContent":               "#!/bin/bash\necho 'Maximal'",
		"runAsAccount":                "user",
		"blockExecutionNotifications": true,
		"enforceSignatureCheck":       true,
	}

	maximalAssignments := []map[string]interface{}{
		{
			"id": "assignment-001",
			"target": map[string]interface{}{
				"@odata.type": "#microsoft.graph.allDevicesAssignmentTarget",
			},
		},
		{
			"id": "assignment-002",
			"target": map[string]interface{}{
				"@odata.type": "#microsoft.graph.groupAssignmentTarget",
				"groupId":     "44444444-4444-4444-4444-444444444444",
			},
		},
	}

	mockState.Lock()
	mockState.scripts[maximalScriptId] = maximalScriptData
	mockState.assignments[maximalScriptId] = maximalAssignments
	mockState.Unlock()
}
