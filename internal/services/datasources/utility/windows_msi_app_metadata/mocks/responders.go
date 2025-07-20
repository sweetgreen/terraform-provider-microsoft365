package mocks

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/jarcoal/httpmock"
)

// WindowsMSIAppMetadataMock provides mock responses for Windows MSI app metadata operations
type WindowsMSIAppMetadataMock struct{}

// RegisterMocks registers HTTP mock responses for Windows MSI app metadata operations
func (m *WindowsMSIAppMetadataMock) RegisterMocks() {
	// Mock Firefox MSI download URL
	firefoxMSIURL := "https://download.mozilla.org/?product=firefox-msi-latest-ssl&os=win64&lang=en-US"
	
	// Register responder for Firefox MSI download
	httpmock.RegisterResponder("GET", firefoxMSIURL,
		func(req *http.Request) (*http.Response, error) {
			// Return a minimal valid MSI file (just the header) for testing
			// This is a valid MSI header that the extraction code can parse
			msiHeader := []byte{
				0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1, // OLE signature
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // CLSID
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Reserved
				0x3E, 0x00, 0x03, 0x00, 0xFE, 0xFF, 0x09, 0x00, // Minor version, DLL version, byte order, sector size
			}
			
			// Pad to make it a realistic size for testing (1KB)
			mockMSIData := make([]byte, 1024)
			copy(mockMSIData, msiHeader)
			
			// Create a response with proper MSI content type and disposition headers
			resp := httpmock.NewBytesResponse(200, mockMSIData)
			resp.Header.Set("Content-Type", "application/x-msi")
			resp.Header.Set("Content-Disposition", "attachment; filename=Firefox_Setup_Test.msi")
			resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(mockMSIData)))
			
			return resp, nil
		})

	// Register responder for URL with redirects (to test redirect handling)
	httpmock.RegisterResponder("GET", `=~^https://download\.mozilla\.org/`,
		func(req *http.Request) (*http.Response, error) {
			// Check if this is already the final URL
			if req.URL.Query().Get("product") == "firefox-msi-latest-ssl" {
				return m.createMockMSIResponse()
			}
			
			// Simulate a redirect to the actual download URL
			resp := &http.Response{
				StatusCode: 302,
				Header: http.Header{
					"Location": []string{firefoxMSIURL},
				},
			}
			return resp, nil
		})

	// Register a fallback responder for any other download URLs
	httpmock.RegisterResponder("GET", `=~^https://.*\.msi$`,
		func(req *http.Request) (*http.Response, error) {
			return m.createMockMSIResponse()
		})
}

// RegisterErrorMocks registers HTTP mock responses for error scenarios
func (m *WindowsMSIAppMetadataMock) RegisterErrorMocks() {
	// Register responder that returns network errors
	httpmock.RegisterResponder("GET", `=~^https://download\.mozilla\.org/`,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(404, "File not found"), nil
		})
	
	// Register responder for timeout scenarios
	httpmock.RegisterResponder("GET", `=~^https://.*timeout.*`,
		func(req *http.Request) (*http.Response, error) {
			return nil, &url.Error{
				Op:  "Get",
				URL: req.URL.String(),
				Err: fmt.Errorf("timeout"),
			}
		})
}

// createMockMSIResponse creates a mock MSI file response
func (m *WindowsMSIAppMetadataMock) createMockMSIResponse() (*http.Response, error) {
	// Create a minimal valid MSI file structure for testing
	// This includes the basic OLE compound document header and some MSI-specific structures
	msiData := []byte{
		// OLE Compound Document header
		0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1, // Signature
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // CLSID
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Reserved
		0x3E, 0x00, 0x03, 0x00, 0xFE, 0xFF, 0x09, 0x00, // Minor version, DLL version, byte order, sector size
		0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Major version, reserved
	}
	
	// Pad to make it larger (simulate a small MSI file - 2KB)
	mockMSIData := make([]byte, 2048)
	copy(mockMSIData, msiData)
	
	// Add some mock MSI property data (simplified)
	mockProperties := []byte{
		// Mock property table entries
		'P', 'r', 'o', 'd', 'u', 'c', 't', 'N', 'a', 'm', 'e', 0x00, // ProductName
		'M', 'o', 'z', 'i', 'l', 'l', 'a', ' ', 'F', 'i', 'r', 'e', 'f', 'o', 'x', 0x00, // Mozilla Firefox
		'P', 'r', 'o', 'd', 'u', 'c', 't', 'V', 'e', 'r', 's', 'i', 'o', 'n', 0x00, // ProductVersion
		'1', '4', '0', '.', '0', '.', '4', 0x00, // 140.0.4
		'P', 'r', 'o', 'd', 'u', 'c', 't', 'C', 'o', 'd', 'e', 0x00, // ProductCode
		'{', '1', '2', '3', '4', '5', '6', '7', '8', '-', 'A', 'B', 'C', 'D', '-', '1', '2', '3', '4', '-', 'A', 'B', 'C', 'D', '-', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '}', 0x00,
		'M', 'a', 'n', 'u', 'f', 'a', 'c', 't', 'u', 'r', 'e', 'r', 0x00, // Manufacturer
		'M', 'o', 'z', 'i', 'l', 'l', 'a', 0x00, // Mozilla
	}
	
	// Copy mock properties to offset 512 in the file
	if len(mockMSIData) > 512+len(mockProperties) {
		copy(mockMSIData[512:], mockProperties)
	}
	
	resp := httpmock.NewBytesResponse(200, mockMSIData)
	resp.Header.Set("Content-Type", "application/x-msi")
	resp.Header.Set("Content-Disposition", "attachment; filename=Firefox_Setup_Test.msi")
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(mockMSIData)))
	
	return resp, nil
}