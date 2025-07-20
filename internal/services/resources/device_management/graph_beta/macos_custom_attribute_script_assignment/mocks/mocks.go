package mocks

// RegisterMocks registers all mock responders for the macOS custom attribute script assignment resource
func RegisterMocks() {
	mock := &MacosCustomAttributeScriptAssignmentMock{}
	mock.RegisterMocks()
	mock.RegisterErrorMocks()
}
