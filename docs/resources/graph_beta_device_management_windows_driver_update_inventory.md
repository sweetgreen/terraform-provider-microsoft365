---
page_title: "microsoft365_graph_beta_device_management_windows_driver_update_inventory Resource - microsoft365"
subcategory: "Device Management"

description: |-
  Manages Windows Driver Update Inventory in Microsoft Intune.
---

# microsoft365_graph_beta_device_management_windows_driver_update_inventory (Resource)

Manages Windows Driver Update Inventory in Microsoft Intune.

## Microsoft Documentation

- [windowsDriverUpdateInventory resource type](https://learn.microsoft.com/en-us/graph/api/resources/intune-softwareupdate-windowsdriverupdateinventory?view=graph-rest-beta)
- [Create windowsDriverUpdateInventory](https://learn.microsoft.com/en-us/graph/api/intune-softwareupdate-windowsdriverupdateinventory-create?view=graph-rest-beta)

## API Permissions

The following API permissions are required in order to use this resource.

### Microsoft Graph

- **Application**: `DeviceManagementConfiguration.ReadWrite.All`

## Version History

| Version | Status | Notes |
|---------|--------|-------|
| v0.14.1-alpha | Experimental | Initial release |

## Example Usage

```terraform
resource "microsoft365_graph_beta_device_management_windows_driver_update_inventory" "example" {
  # Required attributes
  name                             = "Intel HD Graphics Driver"
  version                          = "27.20.100.8681"
  manufacturer                     = "Intel Corporation"
  approval_status                  = "approved"                             # Possible values: "needsReview", "declined", "approved", "suspended"
  category                         = "recommended"                          # Possible values: "recommended", "previouslyApproved", "other"
  windows_driver_update_profile_id = "12345678-1234-1234-1234-123456789012" # ID of the Windows Driver Update Profile

  # Optional attributes
  release_date_time = "2024-12-15T00:00:00Z"
  driver_class      = "Display"
  deploy_date_time  = "2025-01-15T00:00:00Z" # Only needed if approval_status is "approved"

  # Optional timeouts
  timeouts = {
    create = "3m"
    update = "3m"
    read   = "3m"
    delete = "3m"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `approval_status` (String) The approval status for this driver. Possible values are: `needsReview`, `declined`, `approved`, `suspended`.
- `category` (String) The category for this driver. Possible values are: `recommended`, `previouslyApproved`, `other`.
- `manufacturer` (String) The manufacturer of the driver.
- `name` (String) The name of the driver.
- `version` (String) The version of the driver.
- `windows_driver_update_profile_id` (String) The ID of the Windows Driver Update Profile this inventory belongs to.

### Optional

- `deploy_date_time` (String) The date time when a driver should be deployed if approvalStatus is approved.
- `driver_class` (String) The class of the driver.
- `release_date_time` (String) The release date time of the driver.
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `applicable_device_count` (Number) The number of devices for which this driver is applicable.
- `id` (String) The id of the driver.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

Import is supported using the following syntax:

```shell
# {resource_id}
terraform import microsoft365_graph_beta_device_and_app_management_windows_driver_update_inventory.example windows-driver-update-inventory-id
```