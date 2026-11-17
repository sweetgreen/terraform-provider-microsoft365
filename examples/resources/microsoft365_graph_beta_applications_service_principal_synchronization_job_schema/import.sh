# {service_principal_id}/{synchronization_job_id}/{synchronization_rule_id}
terraform import microsoft365_graph_beta_applications_service_principal_synchronization_job_schema.servicenow \
  "00000000-0000-0000-0000-000000000000/ServiceNowOutDelta.00000000000000000000000000000000.11111111-1111-1111-1111-111111111111/22222222-2222-2222-2222-222222222222"

# The rule ID can be left out when the schema has a single synchronization rule.
terraform import microsoft365_graph_beta_applications_service_principal_synchronization_job_schema.servicenow \
  "00000000-0000-0000-0000-000000000000/ServiceNowOutDelta.00000000000000000000000000000000.11111111-1111-1111-1111-111111111111"
