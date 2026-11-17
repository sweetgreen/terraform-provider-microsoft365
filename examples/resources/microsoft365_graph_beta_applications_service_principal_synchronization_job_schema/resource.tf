# Attribute mappings of a ServiceNow provisioning job.
#
# Import the live schema first so the first plan shows the real difference:
#   GET /servicePrincipals/{id}/synchronization/jobs             -> synchronization_job_id
#   GET /servicePrincipals/{id}/synchronization/jobs/{id}/schema -> synchronizationRules[].id
resource "microsoft365_graph_beta_applications_service_principal_synchronization_job_schema" "servicenow" {
  service_principal_id    = "00000000-0000-0000-0000-000000000000"
  synchronization_job_id  = "ServiceNowOutDelta.00000000000000000000000000000000.11111111-1111-1111-1111-111111111111"
  synchronization_rule_id = "22222222-2222-2222-2222-222222222222"

  object_mappings = {
    sys_user = {
      name               = "Provision Microsoft Entra ID Users"
      enabled            = true
      source_object_name = "User"
      flow_types         = ["Add", "Update", "Delete"]

      attribute_mappings = {
        # Deprovisioned (soft-deleted) users become inactive in ServiceNow.
        active = {
          source_expression = "Switch([IsSoftDeleted], , \"False\", \"1\", \"True\", \"0\")"
          default_value     = "1"
        }
        user_name = {
          source_expression = "[userPrincipalName]"
          matching_priority = 1
        }
        email      = { source_expression = "[mail]" }
        first_name = { source_expression = "[givenName]" }
        last_name  = { source_expression = "[surname]" }
        title      = { source_expression = "[jobTitle]" }
        location = {
          source_expression = "[physicalDeliveryOfficeName]"
          default_value     = ""
        }
      }
    }

    sys_user_group = {
      name               = "Provision Microsoft Entra ID Groups"
      enabled            = true
      source_object_name = "Group"
      flow_types         = ["Add", "Update", "Delete"]

      # Only security groups.
      scope = jsonencode({
        categoryFilterGroups = []
        inputFilterGroups    = []
        groups = [{
          name = "security groups"
          clauses = [{
            operatorName      = "IsTrue"
            sourceOperandName = "securityEnabled"
            targetOperand     = null
          }]
        }]
      })

      attribute_mappings = {
        active = {
          source_expression = "Switch([IsSoftDeleted], , \"False\", \"1\", \"True\", \"0\")"
          default_value     = "1"
        }
        name = {
          source_expression = "[displayName]"
          matching_priority = 1
        }
        description = { source_expression = "[description]" }
        email       = { source_expression = "[mail]" }
        user        = { source_expression = "[members]" }
      }
    }

    # A template mapping that is switched off. Attribute mappings without a
    # source_expression have no source in Entra.
    cmn_department = {
      enabled            = false
      source_object_name = ""
      flow_types         = ["Add", "Update", "Delete"]
      attribute_mappings = {
        name = { matching_priority = 1 }
      }
    }
  }
}
