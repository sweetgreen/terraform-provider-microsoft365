package mocks

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/jarcoal/httpmock"

	common "github.com/deploymenttheory/terraform-provider-microsoft365/internal/mocks"
)

const (
	ServicePrincipalID = "11111111-1111-1111-1111-111111111111"
	JobID              = "ServiceNowOutDelta.00000000000000000000000000000000.22222222-2222-2222-2222-222222222222"
	RuleID             = "33333333-3333-3333-3333-333333333333"

	schemaURL = "https://graph.microsoft.com/beta/servicePrincipals/" + ServicePrincipalID +
		"/synchronization/jobs/" + JobID + "/schema"
)

// Mock serves one provisioning job's schema from tests/responses/get_schema.json.
type Mock struct {
	sync.Mutex
	// Schema is the stored schema document.
	Schema map[string]any
	// Puts holds every PUT body received.
	Puts []map[string]any
	// FunctionGets counts GET .../schema/functions requests.
	FunctionGets int
	// pendingStaleReads arms staleReads at the next PUT; staleReads is the number of
	// GETs that still return stale, the schema from before that PUT.
	pendingStaleReads int
	staleReads        int
	stale             map[string]any
	// changeVersionOnEveryGet simulates someone editing the schema in the portal
	// between every read.
	changeVersionOnEveryGet bool
	gets                    int
}

var _ common.MockRegistrar = (*Mock)(nil)

func init() {
	common.GlobalRegistry.Register("service_principal_synchronization_job_schema", &Mock{})
}

func load(path string) map[string]any {
	raw, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		panic(err)
	}
	return doc
}

func decode(q *http.Request) (map[string]any, error) {
	var reader io.Reader = q.Body
	if q.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(q.Body)
		if err != nil {
			return nil, fmt.Errorf("decompress request: %w", err)
		}
		defer gz.Close()
		reader = gz
	}
	var body map[string]any
	if err := json.NewDecoder(reader).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode request: %w", err)
	}
	return body, nil
}

// SetStaleReadsAfterPut makes the n GETs after the next PUT return the schema from before it.
func (m *Mock) SetStaleReadsAfterPut(n int) {
	m.Lock()
	defer m.Unlock()
	m.pendingStaleReads = n
}

// ChangeVersionOnEveryGet makes every GET return a new schema version.
func (m *Mock) ChangeVersionOnEveryGet() {
	m.Lock()
	defer m.Unlock()
	m.changeVersionOnEveryGet = true
}

func (m *Mock) RegisterMocks() {
	m.CleanupMockState()
	m.Lock()
	m.Schema = load("tests/responses/get_schema.json")
	m.Unlock()
	functions := load("tests/responses/get_functions.json")

	httpmock.RegisterResponder("GET", schemaURL, func(q *http.Request) (*http.Response, error) {
		m.Lock()
		defer m.Unlock()
		m.gets++
		if m.changeVersionOnEveryGet {
			m.Schema["version"] = fmt.Sprintf("Date:2026-11-17T00:00:%02d.0000000Z, ActivityId:edited-in-portal", m.gets)
		}
		if m.staleReads > 0 {
			m.staleReads--
			return httpmock.NewJsonResponse(200, m.stale)
		}
		return httpmock.NewJsonResponse(200, m.Schema)
	})
	httpmock.RegisterResponder("GET", schemaURL+"/functions", func(q *http.Request) (*http.Response, error) {
		m.Lock()
		defer m.Unlock()
		m.FunctionGets++
		return httpmock.NewJsonResponse(200, functions)
	})
	httpmock.RegisterResponder("PUT", schemaURL, func(q *http.Request) (*http.Response, error) {
		body, err := decode(q)
		if err != nil {
			return nil, err
		}
		for k := range body {
			if strings.Contains(k, "@odata.") {
				return httpmock.NewStringResponse(400,
					`{"error":{"code":"BadRequest","message":"annotation `+k+` is not allowed"}}`), nil
			}
		}
		m.Lock()
		defer m.Unlock()
		m.Puts = append(m.Puts, body)
		m.stale = m.Schema
		m.staleReads = m.pendingStaleReads
		m.pendingStaleReads = 0
		stored := map[string]any{"@odata.context": m.Schema["@odata.context"]}
		for k, v := range body {
			stored[k] = v
		}
		stored["version"] = fmt.Sprintf("Date:2026-11-17T00:00:00.0000000Z, ActivityId:put-%d", len(m.Puts))
		m.Schema = stored
		return httpmock.NewStringResponse(204, ""), nil
	})
}

func (m *Mock) RegisterErrorMocks() {
	for _, method := range []string{"GET", "PUT"} {
		httpmock.RegisterResponder(method, schemaURL, httpmock.NewStringResponder(403,
			`{"error":{"code":"Forbidden","message":"Synthetic denied request"}}`))
	}
}

func (m *Mock) CleanupMockState() {
	m.Lock()
	defer m.Unlock()
	m.Schema = nil
	m.Puts = nil
	m.FunctionGets = 0
	m.pendingStaleReads = 0
	m.staleReads = 0
	m.stale = nil
	m.changeVersionOnEveryGet = false
	m.gets = 0
}
