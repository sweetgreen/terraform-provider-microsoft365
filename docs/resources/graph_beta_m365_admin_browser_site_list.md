---
page_title: "microsoft365_graph_beta_m365_admin_browser_site_list Resource - microsoft365"
subcategory: "M365 Admin"
description: |-
  Manages Internet Explorer mode site lists in Microsoft Edge using the /admin/edge/internetExplorerMode/siteLists endpoint. Site lists are collections of websites that require Internet Explorer 11 compatibility mode, allowing organizations to maintain legacy web applications while transitioning to Microsoft Edge as the default browser.
---

# microsoft365_graph_beta_m365_admin_browser_site_list (Resource)

Manages Internet Explorer mode site lists in Microsoft Edge using the `/admin/edge/internetExplorerMode/siteLists` endpoint. Site lists are collections of websites that require Internet Explorer 11 compatibility mode, allowing organizations to maintain legacy web applications while transitioning to Microsoft Edge as the default browser.

## Microsoft Documentation

- [browserSiteList resource type](https://learn.microsoft.com/en-us/graph/api/resources/browsersitelist?view=graph-rest-beta)

## API Permissions

The following API permissions are required in order to use this resource.

### Microsoft Graph

- **Application**: `BrowserSiteLists.ReadWrite.All`

## Version History

| Version | Status | Notes |
|---------|--------|-------|
| v0.15.0-alpha | Experimental | Initial release

## Example Usage

```terraform
resource "microsoft365_graph_beta_device_and_app_management_browser_site_list" "example" {
  display_name = "Example Browser Site List"
  description  = "This is an example browser site list for Internet Explorer Mode"

  # Optional: Define custom timeouts
  timeouts = {
    create = "30m"
    read   = "10m"
    update = "30m"
    delete = "30m"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `display_name` (String) The name of the site list.

### Optional

- `description` (String) The description of the site list.
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) The unique identifier for the site list.
- `last_modified_date_time` (String) The date and time when the site list was last modified.
- `published_date_time` (String) The date and time when the site list was published.
- `revision` (String) The current revision of the site list.
- `status` (String) The current status of the site list.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Important Notes

- **Browser Site Lists**: This resource manages lists of websites for Microsoft Edge browser policies in Microsoft 365.
- **Enterprise Mode**: Site lists are commonly used to configure Internet Explorer mode sites for Microsoft Edge.
- **Compatibility**: Helps manage legacy web applications that require Internet Explorer for compatibility.
- **Centralized Management**: Provides centralized control over browser behavior for specific websites across the organization.
- **Policy Integration**: Site lists integrate with Microsoft Edge administrative templates and browser policies.
- **URL Patterns**: Supports various URL pattern formats including wildcards and specific domains.

## Import

Import is supported using the following syntax:

```shell
# {resource_id}
terraform import microsoft365_graph_beta_m365_admin_browser_site_list.example browser-site-list-id
```