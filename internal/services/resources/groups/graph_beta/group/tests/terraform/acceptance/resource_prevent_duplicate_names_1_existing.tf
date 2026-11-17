resource "random_string" "group_suffix" {
  length  = 8
  special = false
}

# A group that already exists in the tenant, created without the guard.
resource "microsoft365_graph_beta_groups_group" "existing" {
  display_name     = "acc-test-dup-group-${random_string.group_suffix.result}"
  description      = "existing group"
  mail_nickname    = "acctestdup${random_string.group_suffix.result}"
  mail_enabled     = false
  security_enabled = true
  hard_delete      = true
}
