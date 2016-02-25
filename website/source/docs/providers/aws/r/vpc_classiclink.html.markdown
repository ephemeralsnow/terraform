---
layout: "aws"
page_title: "AWS: aws_vpc_classiclink"
sidebar_current: "docs-aws-resource-vpc-classiclink"
description: |-
  Provides a VPC ClassicLink resource.
---

# aws\_vpc\_classiclink

Provides a VPC ClassicLink resource.

## Example Usage

Basic usage:

```
resource "aws_vpc_classiclink" "web" {
    instance_id = "${aws_instance.web.id}"
    vpc_id = "${aws_vpc.main.id}"
    vpc_security_group_ids = ["${aws_security_group.web.id}"]
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) The ID of an EC2-Classic instance to link to the ClassicLink-enabled VPC.
* `vpc_id` - (Required) The ID of a ClassicLink-enabled VPC.
* `vpc_security_group_ids` - (Required) The ID of one or more of the VPC's security groups. You cannot specify security groups from a different VPC.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the VPC classiclink.
