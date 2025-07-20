# Minimal configuration for macOS custom attribute script assignment
resource "microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment" "test_minimal" {
  macos_custom_attribute_script_id = "00000000-0000-0000-0000-000000000002"

  target = {
    target_type = "allDevices"
  }
}