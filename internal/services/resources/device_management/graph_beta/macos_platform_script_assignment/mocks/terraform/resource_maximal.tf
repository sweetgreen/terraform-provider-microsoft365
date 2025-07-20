resource "microsoft365_graph_beta_device_management_macos_platform_script_assignment" "test_maximal" {
  macos_platform_script_id = "00000000-0000-0000-0000-000000000004"
  target = {
    target_type                                      = "groupAssignment"
    group_id                                         = "44444444-4444-4444-4444-444444444444"
    device_and_app_management_assignment_filter_id   = "55555555-5555-5555-5555-555555555555"
    device_and_app_management_assignment_filter_type = "include"
  }

  timeouts = {
    create = "30m"
    read   = "10m"
    update = "30m"
    delete = "10m"
  }
}