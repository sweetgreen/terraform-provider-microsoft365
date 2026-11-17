resource "random_string" "group_suffix" {
  length  = 8
  special = false
}

resource "microsoft365_graph_beta_groups_group" "existing" {
  display_name     = "acc-test-dup-group-${random_string.group_suffix.result}"
  description      = "existing group"
  mail_nickname    = "acctestdup${random_string.group_suffix.result}"
  mail_enabled     = false
  security_enabled = true
  hard_delete      = true
}

# A second group with the same display name. With the guard on, planning it must fail.
resource "microsoft365_graph_beta_groups_group" "duplicate" {
  display_name            = microsoft365_graph_beta_groups_group.existing.display_name
  description             = "duplicate group"
  mail_nickname           = "acctestdup2${random_string.group_suffix.result}"
  mail_enabled            = false
  security_enabled        = true
  hard_delete             = true
  prevent_duplicate_names = true
}
