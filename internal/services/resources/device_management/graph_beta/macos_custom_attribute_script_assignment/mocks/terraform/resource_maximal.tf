# Maximal configuration for macOS custom attribute script assignment
resource "microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment" "test_maximal" {
  macos_custom_attribute_script_id = "00000000-0000-0000-0000-000000000004"

  target = {
    target_type                                      = "groupAssignment"
    group_id                                         = "22222222-2222-2222-2222-222222222222"
    device_and_app_management_assignment_filter_id   = "33333333-3333-3333-3333-333333333333"
    device_and_app_management_assignment_filter_type = "include"
  }

  timeouts = {
    create = "30m"
    read   = "5m"
    update = "30m"
    delete = "30m"
  }
}

# Additional test resource with all licensed users target
resource "microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment" "test_all_licensed_users" {
  macos_custom_attribute_script_id = "00000000-0000-0000-0000-000000000002"

  target = {
    target_type = "allLicensedUsers"
  }
}

# Additional test resource with exclude filter
resource "microsoft365_graph_beta_device_management_macos_custom_attribute_script_assignment" "test_exclude_filter" {
  macos_custom_attribute_script_id = "00000000-0000-0000-0000-000000000003"

  target = {
    target_type                                      = "groupAssignment"
    group_id                                         = "55555555-5555-5555-5555-555555555555"
    device_and_app_management_assignment_filter_id   = "66666666-6666-6666-6666-666666666666"
    device_and_app_management_assignment_filter_type = "exclude"
  }
}