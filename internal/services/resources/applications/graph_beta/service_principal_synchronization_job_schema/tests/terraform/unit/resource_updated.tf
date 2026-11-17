resource "microsoft365_graph_beta_applications_service_principal_synchronization_job_schema" "test" {
  service_principal_id    = "11111111-1111-1111-1111-111111111111"
  synchronization_job_id  = "ServiceNowOutDelta.00000000000000000000000000000000.22222222-2222-2222-2222-222222222222"
  synchronization_rule_id = "33333333-3333-3333-3333-333333333333"

  object_mappings = {
    sys_user = {
      name               = "Provision Microsoft Entra ID Users"
      enabled            = true
      source_object_name = "User"
      flow_types         = ["Add", "Update", "Delete"]
      attribute_mappings = {
        active = {
          source_expression = "Switch([IsSoftDeleted], \"1\", \"True\", \"0\")"
          default_value     = "0"
        }
        email = {
          source_expression = "Coalesce([mail], [userPrincipalName])"
        }
        user_name = {
          source_expression = "[userPrincipalName]"
          matching_priority = 1
        }
        cost_center = {
          source_expression = "[extension_00000000000000000000000000000000_costCenter]"
          default_value     = ""
        }
        title = {
          source_expression = "[jobTitle]"
          flow_type         = "ObjectAddOnly"
        }
      }
    }
    sys_user_group = {
      name               = "Provision Microsoft Entra ID Groups"
      enabled            = true
      source_object_name = "Group"
      flow_types         = ["Add", "Update", "Delete"]
      attribute_mappings = {
        active = {
          source_expression = "Switch([IsSoftDeleted], , \"False\", \"1\", \"True\", \"0\")"
          default_value     = "1"
        }
        description = { source_expression = "[description]" }
        email       = { source_expression = "[mail]" }
        user        = { source_expression = "[members]" }
        name = {
          source_expression = "[displayName]"
          matching_priority = 1
        }
      }
    }
  }
}
