# Hetzner Cloud Terraform Provider

## Current State
**This project is in alpha state**. Many functions are tested, but their is no guarantee they are. There are espacially no automated tests

### ToDo
 - Implement ISOs
 - Implement Floating IPs
 - Implement IPv6 DNS PTR

## Installation
 - Download the source code
 - run `go build` in the source code folder
 - tell terraform where to find your binary (in your project, or in ~/.terraformrc)
```
providers {
    hcloud = "/path/to/folder/terraform-provider-hcloud"
}
```

## Configuration
 - create a project in your hetzner cloud console
 - create a access token for this project
 - terraform will either prompt you for the token or you can specify it in your project
```
provider "hcloud" {
    token = "YOUR-TOKEN"
}
```

## Data Sources

### Server Type
```
data "hcloud_servertype" "cx11" {
    name = "cx11"
}
```
Using the standard name will select the respective server type with local storage. To use network storage add '-ceph', e.g. 'cx11-ceph'.

*Output:* id

### Image
```
data "hcloud_image" "debian" {
    name = "debian-9"
}
```
*Output:* id

### Datacenter
```
data "hcloud_datacenter" "fsn1-dc8" {
    name = "fsn1-dc8"
}
```

*Output*: id

### Location
```
data "hcloud_location" "falkenstein" {
    name = "fsn1"
}
```

*Output:* id

## Resources

### Server
```
resource "hcloud_server" "test" {
    name = "testserver"                                     // String, required
    server_type = "${data.hcloud_servertype.cx11.id}"       // Int, required
    datacenter = "${data.hcloud_datacenter.fsn1-dc8.id}"    // Int, optional (conflicts with location)
    location = "${data.hcloud_location.falkenstein.id}"     // Int, optional
    image = "${data.hcloud_image.debian.id}"                // Int, required
    ssh_keys = [                                            // []Int, optional
        "${hcloud_sshkey.my-key.id}"
    ]
    user_data = ""                                          // String, optional
    ipv4_ptr = ""                                           // String, optional
    upgrade_disk = "true"                                   // Bool, optional (std: true)
    backup = "false"                                        // Bool, optional (std: false)
    backup_window = ""                                      // String, optional
}
```

*Outputs:* datacenter (Int), location (Int), status (String), created (String), ipv4 (String), ipv6 (String), ipv6_ptr (map[ip String][ptr String]), root_pw (String), backup_window (String)

- Setting the IPv6 DNS PTR is currently not supported
- To enable backups set `backup = "true"` if 'backup_window' is empty a random time slot will be assigned
- 'upgrade_disk' is only used if you change the server type

### SSHKey
```
resource "hcloud_sshkey" "my-key" {
    name = "Kajos MacBook Pro"
    public_key = "ssh-rsa AAAAB3N....."
}
```

*Output:* fingerprint, id

### Rescue
```
resource "hcloud_rescue" "test" {
    server = "${hcloud_server.test.id}"                 // Int, required
    type = "linux64"                                    // String, optional
    ssh_keys = [                                        // []Int, optional
        "${hcloud_sshkey.my-key.id}"
    ]
    reset_on_activation = "true"                        // Bool, optional (Std: true)
    reboot_on_activation = "false"                      // Bool, optional (Std: false)
    reset_on_deactivation = "true"                      // Bool, optional (Std: true)
    reboot_on_deactivation = "false"                    // Bool, optional (Std: false)
}
```

*Outputs:* password

 - if 'reboot_on_...' is true, the provider will initiate a soft reboot on (de)activation. Make sure your system has ACPI support, if you use this option
 - if 'reboot_on_...' is false and 'reset_on_...' is true, the provider will initiate a reset on (de)activation
 - if both are false, the provider will just (de)activate the rescue system for the next boot
 
 
This resource will generate a uuid as id and will not check if the rescue system is still running. The main purpose of this resource is to boot the server into the rescue system and then provision it. If you want to reprovision the server, just taint the resource. Whene deleting this resource it will make sure to disable the rescue system for the next boot and depending on your settings reboot the server.