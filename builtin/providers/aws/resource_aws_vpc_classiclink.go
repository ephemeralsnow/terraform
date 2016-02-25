package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsVpcClassicLink() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVPCClassicLinkCreate,
		Read:   resourceAwsVPCClassicLinkRead,
		Update: resourceAwsVPCClassicLinkUpdate,
		Delete: resourceAwsVPCClassicLinkDelete,
		Schema: map[string]*schema.Schema{
			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_security_group_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func resourceAwsVPCClassicLinkCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	input := &ec2.AttachClassicLinkVpcInput{
		InstanceId: aws.String(d.Get("instance_id").(string)),
		VpcId:      aws.String(d.Get("vpc_id").(string)),
		Groups:     expandStringList(d.Get("vpc_security_group_ids").(*schema.Set).List()),
	}

	log.Printf("[DEBUG] Creating VPC ClassicLink: %#v", input)
	_, err := conn.AttachClassicLinkVpc(input)
	if err != nil {
		return fmt.Errorf("Error creating VPC ClassicLink: %s", err)
	}
	log.Printf("[DEBUG] VPC ClassicLink %q to %q created.", *input.InstanceId, *input.VpcId)

	d.SetId(*input.InstanceId + "/" + *input.VpcId)

	return resourceAwsVPCClassicLinkRead(d, meta)
}

func resourceAwsVPCClassicLinkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	instanceId := d.Get("instance_id").(string)
	input := &ec2.DescribeClassicLinkInstancesInput{
		InstanceIds: []*string{aws.String(instanceId)},
	}

	log.Printf("[DEBUG] Reading VPC ClassicLink: %q", instanceId)
	output, err := conn.DescribeClassicLinkInstances(input)

	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return fmt.Errorf("Error reading VPC ClassicLink: %s", err.Error())
		}

		if ec2err.Code() == "InvalidInstanceID.NotFound" {
			return nil
		}

		return fmt.Errorf("Error reading VPC ClassicLink: %s", err.Error())
	}

	if len(output.Instances) == 0 {
		return nil
	}

	instance := output.Instances[0]

	d.Set("instance_id", *instance.InstanceId)
	d.Set("vpc_id", *instance.VpcId)

	sgs := make([]string, 0, len(instance.Groups))
	for _, sg := range instance.Groups {
		sgs = append(sgs, *sg.GroupId)
	}
	log.Printf("[DEBUG] Setting Security Group IDs: %#v", sgs)
	if err := d.Set("vpc_security_group_ids", sgs); err != nil {
		return err
	}

	return nil
}

func resourceAwsVPCClassicLinkUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("vpc_id") || d.HasChange("vpc_security_group_ids") {
		if err := resourceAwsVPCClassicLinkDelete(d, meta); err != nil {
			return err
		}
		if err := resourceAwsVPCClassicLinkCreate(d, meta); err != nil {
			return err
		}
	}

	return resourceAwsVPCClassicLinkRead(d, meta)
}

func resourceAwsVPCClassicLinkDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	input := &ec2.DetachClassicLinkVpcInput{
		InstanceId: aws.String(d.Get("instance_id").(string)),
		VpcId:      aws.String(d.Get("vpc_id").(string)),
	}

	log.Printf("[DEBUG] Deleting VPC ClassicLink: %#v", input)
	_, err := conn.DetachClassicLinkVpc(input)

	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return fmt.Errorf("Error deleting VPC ClassicLink: %s", err.Error())
		}

		if ec2err.Code() == "InvalidInstanceID.NotFound" {
			log.Printf("[DEBUG] VPC ClassicLink %q is already gone", d.Id())
		} else {
			return fmt.Errorf("Error deleting VPC ClassicLink: %s", err.Error())
		}
	}

	log.Printf("[DEBUG] VPC ClassicLink %q deleted", d.Id())
	d.SetId("")

	return nil
}
