package common

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

var reShutdownBehavior = regexp.MustCompile("^(stop|terminate)$")

type AmiFilterOptions struct {
	Filters    map[*string]*string
	Owners     []*string
	MostRecent bool `mapstructure:"most_recent"`
}

func (d *AmiFilterOptions) Empty() bool {
	return len(d.Owners) == 0 && len(d.Filters) == 0
}

func (d *AmiFilterOptions) NoOwner() bool {
	return len(d.Owners) == 0
}

type SubnetFilterOptions struct {
	Filters  map[*string]*string
	MostFree bool `mapstructure:"most_free"`
	Random   bool `mapstructure:"random"`
}

func (d *SubnetFilterOptions) Empty() bool {
	return len(d.Filters) == 0
}

type VpcFilterOptions struct {
	Filters map[*string]*string
}

func (d *VpcFilterOptions) Empty() bool {
	return len(d.Filters) == 0
}

type SecurityGroupFilterOptions struct {
	Filters map[*string]*string
}

func (d *SecurityGroupFilterOptions) Empty() bool {
	return len(d.Filters) == 0
}

// RunConfig contains configuration for running an instance from a source
// AMI and details on how to access that launched image.
type RunConfig struct {
	// If using a non-default VPC, public IP addresses are not provided by default. If this is true, your new instance will get a Public IP. default: false
	AssociatePublicIpAddress          bool                       `mapstructure:"associate_public_ip_address" required:"false" required:"false" required:"false" required:"false"`
	// Destination availability zone to launch instance in. Leave this empty to allow Amazon to auto-assign.
	AvailabilityZone                  string                     `mapstructure:"availability_zone" required:"false" required:"false" required:"false" required:"false"`
	// Requires spot_price to be set. The required duration for the Spot Instances (also known as Spot blocks). This value must be a multiple of 60 (60, 120, 180, 240, 300, or 360). You can't specify an Availability Zone group or a launch group if you specify a duration.
	BlockDurationMinutes              int64                      `mapstructure:"block_duration_minutes" required:"false" required:"false" required:"false" required:"false"`
	// Packer normally stops the build instance after all provisioners have run. For Windows instances, it is sometimes desirable to run Sysprep which will stop the instance for you. If this is set to true, Packer will not stop the instance but will assume that you will send the stop signal yourself through your final provisioner. You can do this with a windows-shell provisioner.
	DisableStopInstance               bool                       `mapstructure:"disable_stop_instance" required:"false" required:"false" required:"false"`
	// Mark instance as EBS Optimized. Default false.
	EbsOptimized                      bool                       `mapstructure:"ebs_optimized" required:"false" required:"false" required:"false" required:"false"`
	// Enabling T2 Unlimited allows the source instance to burst additional CPU beyond its available CPU Credits for as long as the demand exists. This is in contrast to the standard configuration that only allows an instance to consume up to its available CPU Credits. See the AWS documentation for T2 Unlimited and the T2 Unlimited Pricing section of the Amazon EC2 On-Demand Pricing document for more information. By default this option is disabled and Packer will set up a T2 Standard instance instead.
	EnableT2Unlimited                 bool                       `mapstructure:"enable_t2_unlimited" required:"false" required:"false" required:"false" required:"false"`
	// The name of an IAM instance profile to launch the EC2 instance with.
	IamInstanceProfile                string                     `mapstructure:"iam_instance_profile" required:"false" required:"false"`
	// Automatically terminate instances on shutdown in case Packer exits ungracefully. Possible values are stop and terminate. Defaults to stop.
	InstanceInitiatedShutdownBehavior string                     `mapstructure:"shutdown_behavior" required:"false"`
	// The EC2 instance type to use while building the AMI, such as t2.small.
	InstanceType                      string                     `mapstructure:"instance_type" required:"true"`
	// Filters used to populate the security_group_ids field. Example:
	SecurityGroupFilter               SecurityGroupFilterOptions `mapstructure:"security_group_filter" required:"false" required:"false"`
	RunTags                           map[string]string          `mapstructure:"run_tags"`
	// The ID (not the name) of the security group to assign to the instance. By default this is not set and Packer will automatically create a new temporary security group to allow SSH access. Note that if this is specified, you must be sure the security group allows access to the ssh_port given below.
	SecurityGroupId                   string                     `mapstructure:"security_group_id" required:"false" required:"false"`
	SecurityGroupIds                  []string                   `mapstructure:"security_group_ids"`
	// The source AMI whose root volume will be copied and provisioned on the currently running instance. This must be an EBS-backed AMI with a root volume snapshot that you have access to. Note: this is not used when from_scratch is set to true.
	SourceAmi                         string                     `mapstructure:"source_ami" required:"true"`
	// Filters used to populate the source_ami field. Example:
	SourceAmiFilter                   AmiFilterOptions           `mapstructure:"source_ami_filter" required:"false" required:"false"`
	SpotInstanceTypes                 []string                   `mapstructure:"spot_instance_types"`
	SpotPrice                         string                     `mapstructure:"spot_price"`
	SpotPriceAutoProduct              string                     `mapstructure:"spot_price_auto_product"`
	SpotTags                          map[string]string          `mapstructure:"spot_tags"`
	SubnetFilter                      SubnetFilterOptions        `mapstructure:"subnet_filter"`
	SubnetId                          string                     `mapstructure:"subnet_id"`
	TemporaryKeyPairName              string                     `mapstructure:"temporary_key_pair_name"`
	TemporarySGSourceCidrs            []string                   `mapstructure:"temporary_security_group_source_cidrs"`
	UserData                          string                     `mapstructure:"user_data"`
	UserDataFile                      string                     `mapstructure:"user_data_file"`
	VpcFilter                         VpcFilterOptions           `mapstructure:"vpc_filter"`
	VpcId                             string                     `mapstructure:"vpc_id"`
	WindowsPasswordTimeout            time.Duration              `mapstructure:"windows_password_timeout"`

	// Communicator settings
	Comm communicator.Config `mapstructure:",squash"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	// If we are not given an explicit ssh_keypair_name or
	// ssh_private_key_file, then create a temporary one, but only if the
	// temporary_key_pair_name has not been provided and we are not using
	// ssh_password.
	if c.Comm.SSHKeyPairName == "" && c.Comm.SSHTemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKeyFile == "" && c.Comm.SSHPassword == "" {

		c.Comm.SSHTemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	if c.WindowsPasswordTimeout == 0 {
		c.WindowsPasswordTimeout = 20 * time.Minute
	}

	if c.RunTags == nil {
		c.RunTags = make(map[string]string)
	}

	// Validation
	errs := c.Comm.Prepare(ctx)

	// Validating ssh_interface
	if c.Comm.SSHInterface != "public_ip" &&
		c.Comm.SSHInterface != "private_ip" &&
		c.Comm.SSHInterface != "public_dns" &&
		c.Comm.SSHInterface != "private_dns" &&
		c.Comm.SSHInterface != "" {
		errs = append(errs, fmt.Errorf("Unknown interface type: %s", c.Comm.SSHInterface))
	}

	if c.Comm.SSHKeyPairName != "" {
		if c.Comm.Type == "winrm" && c.Comm.WinRMPassword == "" && c.Comm.SSHPrivateKeyFile == "" {
			errs = append(errs, fmt.Errorf("ssh_private_key_file must be provided to retrieve the winrm password when using ssh_keypair_name."))
		} else if c.Comm.SSHPrivateKeyFile == "" && !c.Comm.SSHAgentAuth {
			errs = append(errs, fmt.Errorf("ssh_private_key_file must be provided or ssh_agent_auth enabled when ssh_keypair_name is specified."))
		}
	}

	if c.SourceAmi == "" && c.SourceAmiFilter.Empty() {
		errs = append(errs, fmt.Errorf("A source_ami or source_ami_filter must be specified"))
	}

	if c.SourceAmi == "" && c.SourceAmiFilter.NoOwner() {
		errs = append(errs, fmt.Errorf("For security reasons, your source AMI filter must declare an owner."))
	}

	if c.InstanceType == "" && len(c.SpotInstanceTypes) == 0 {
		errs = append(errs, fmt.Errorf("either instance_type or "+
			"spot_instance_types must be specified"))
	}

	if c.InstanceType != "" && len(c.SpotInstanceTypes) > 0 {
		errs = append(errs, fmt.Errorf("either instance_type or "+
			"spot_instance_types must be specified, not both"))
	}

	if c.BlockDurationMinutes%60 != 0 {
		errs = append(errs, fmt.Errorf(
			"block_duration_minutes must be multiple of 60"))
	}

	if c.SpotPrice == "auto" {
		if c.SpotPriceAutoProduct == "" {
			errs = append(errs, fmt.Errorf(
				"spot_price_auto_product must be specified when spot_price is auto"))
		}
	}

	if c.SpotPriceAutoProduct != "" {
		if c.SpotPrice != "auto" {
			errs = append(errs, fmt.Errorf(
				"spot_price should be set to auto when spot_price_auto_product is specified"))
		}
	}

	if c.SpotTags != nil {
		if c.SpotPrice == "" || c.SpotPrice == "0" {
			errs = append(errs, fmt.Errorf(
				"spot_tags should not be set when not requesting a spot instance"))
		}
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = append(errs, fmt.Errorf("Only one of user_data or user_data_file can be specified."))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = append(errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	if c.SecurityGroupId != "" {
		if len(c.SecurityGroupIds) > 0 {
			errs = append(errs, fmt.Errorf("Only one of security_group_id or security_group_ids can be specified."))
		} else {
			c.SecurityGroupIds = []string{c.SecurityGroupId}
			c.SecurityGroupId = ""
		}
	}

	if len(c.TemporarySGSourceCidrs) == 0 {
		c.TemporarySGSourceCidrs = []string{"0.0.0.0/0"}
	} else {
		for _, cidr := range c.TemporarySGSourceCidrs {
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				errs = append(errs, fmt.Errorf("Error parsing CIDR in temporary_security_group_source_cidrs: %s", err.Error()))
			}
		}
	}

	if c.InstanceInitiatedShutdownBehavior == "" {
		c.InstanceInitiatedShutdownBehavior = "stop"
	} else if !reShutdownBehavior.MatchString(c.InstanceInitiatedShutdownBehavior) {
		errs = append(errs, fmt.Errorf("shutdown_behavior only accepts 'stop' or 'terminate' values."))
	}

	if c.EnableT2Unlimited {
		if c.SpotPrice != "" {
			errs = append(errs, fmt.Errorf("Error: T2 Unlimited cannot be used in conjuction with Spot Instances"))
		}
		firstDotIndex := strings.Index(c.InstanceType, ".")
		if firstDotIndex == -1 {
			errs = append(errs, fmt.Errorf("Error determining main Instance Type from: %s", c.InstanceType))
		} else if c.InstanceType[0:firstDotIndex] != "t2" {
			errs = append(errs, fmt.Errorf("Error: T2 Unlimited enabled with a non-T2 Instance Type: %s", c.InstanceType))
		}
	}

	return errs
}

func (c *RunConfig) IsSpotInstance() bool {
	return c.SpotPrice != "" && c.SpotPrice != "0"
}
