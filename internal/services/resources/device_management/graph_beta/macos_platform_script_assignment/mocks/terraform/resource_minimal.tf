resource "microsoft365_graph_beta_device_management_macos_platform_script_assignment" "test_minimal" {
  macos_platform_script_id = "00000000-0000-0000-0000-000000000003"
  target = {
    target_type = "allDevices"
  }
}