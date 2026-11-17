resource "random_string" "group_suffix" {
  length  = 8
  special = false
}

# Stop managing the group under its old address without deleting it.
removed {
  from = microsoft365_graph_beta_groups_group.existing

  lifecycle {
    destroy = false
  }
}

# Adopt the same group, imported, with the guard on. The check never runs for a resource with
# prior state, so the import and later updates succeed.
resource "microsoft365_graph_beta_groups_group" "adopted" {
  display_name            = "acc-test-dup-group-${random_string.group_suffix.result}"
  description             = "existing group"
  mail_nickname           = "acctestdup${random_string.group_suffix.result}"
  mail_enabled            = false
  security_enabled        = true
  hard_delete             = true
  prevent_duplicate_names = true
}
