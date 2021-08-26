---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "grid_kubernetes Resource - grid-provider"
subcategory: ""
description: |-
  Sample resource in the Terraform provider scaffolding.
---

# grid_kubernetes (Resource)

Sample resource in the Terraform provider scaffolding.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **master** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--master))
- **nodes_ip_range** (Map of String)
- **token** (String) The cluster secret token

### Optional

- **disks** (Block List) (see [below for nested schema](#nestedblock--disks))
- **id** (String) The ID of this resource.
- **network_name** (String)
- **ssh_key** (String)
- **workers** (Block List) (see [below for nested schema](#nestedblock--workers))

### Read-Only

- **node_deployment_id** (Map of Number)

<a id="nestedblock--master"></a>
### Nested Schema for `master`

Required:

- **cpu** (Number) CPU size
- **disk_size** (Number) Data disk size
- **memory** (Number) Memory size
- **name** (String)
- **node** (Number) Node ID

Optional:

- **env_vars** (Block List) (see [below for nested schema](#nestedblock--master--env_vars))
- **flist** (String)
- **ip** (String) IP
- **mounts** (Block List) (see [below for nested schema](#nestedblock--master--mounts))
- **publicip** (Boolean) If you want to enable public ip or not

Read-Only:

- **computedip** (String) The public ip
- **version** (Number) Version

<a id="nestedblock--master--env_vars"></a>
### Nested Schema for `master.env_vars`

Required:

- **key** (String)
- **value** (String)


<a id="nestedblock--master--mounts"></a>
### Nested Schema for `master.mounts`

Required:

- **disk_name** (String)
- **mount_point** (String)



<a id="nestedblock--disks"></a>
### Nested Schema for `disks`

Required:

- **description** (String)
- **name** (String)
- **nodeid** (Number) Node ID
- **size** (Number)

Optional:

- **version** (Number) Version


<a id="nestedblock--workers"></a>
### Nested Schema for `workers`

Required:

- **cpu** (Number) CPU size
- **disk_size** (Number) Data disk size
- **memory** (Number) Memory size
- **name** (String)
- **node** (Number) Node ID

Optional:

- **env_vars** (Block List) (see [below for nested schema](#nestedblock--workers--env_vars))
- **flist** (String)
- **ip** (String) IP
- **mounts** (Block List) (see [below for nested schema](#nestedblock--workers--mounts))
- **publicip** (Boolean) If you want to enable public ip or not
- **version** (Number) Version

Read-Only:

- **computedip** (String) The public ip

<a id="nestedblock--workers--env_vars"></a>
### Nested Schema for `workers.env_vars`

Required:

- **key** (String)
- **value** (String)


<a id="nestedblock--workers--mounts"></a>
### Nested Schema for `workers.mounts`

Required:

- **disk_name** (String)
- **mount_point** (String)

