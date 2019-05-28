package cloudstack

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// Config holds all the details needed to configure the builder.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	common.HTTPConfig   `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	// The CloudStack API endpoint we will connect to. It can also be specified via environment variable CLOUDSTACK_API_URL, if set.

	APIURL       string        `mapstructure:"api_url" required:"true"`
	// The API key used to sign all API requests. It can also be specified via environment variable CLOUDSTACK_API_KEY, if set.
	APIKey       string        `mapstructure:"api_key" required:"true"`
	// The secret key used to sign all API requests. It can also be specified via environment variable CLOUDSTACK_SECRET_KEY, if set.
	SecretKey    string        `mapstructure:"secret_key" required:"true"`
	AsyncTimeout time.Duration `mapstructure:"async_timeout"`
	HTTPGetOnly  bool          `mapstructure:"http_get_only"`
	SSLNoVerify  bool          `mapstructure:"ssl_no_verify"`

	CIDRList               []string `mapstructure:"cidr_list"`
	CreateSecurityGroup    bool     `mapstructure:"create_security_group"`
	DiskOffering           string   `mapstructure:"disk_offering"`
	DiskSize               int64    `mapstructure:"disk_size"`
	Expunge                bool     `mapstructure:"expunge"`
	Hypervisor             string   `mapstructure:"hypervisor"`
	InstanceName           string   `mapstructure:"instance_name"`
	// The name or ID of the network to connect the instance to.
	Network                string   `mapstructure:"network" required:"true"`
	Project                string   `mapstructure:"project"`
	PublicIPAddress        string   `mapstructure:"public_ip_address"`
	PublicPort             int      `mapstructure:"public_port"`
	SecurityGroups         []string `mapstructure:"security_groups"`
	// The name or ID of the service offering used for the instance.
	ServiceOffering        string   `mapstructure:"service_offering" required:"true"`
	PreventFirewallChanges bool     `mapstructure:"prevent_firewall_changes"`
	// The name or ID of an ISO that will be mounted before booting the instance. This option is mutually exclusive with source_template. When using source_iso, both disk_offering and hypervisor are required.
	SourceISO              string   `mapstructure:"source_iso" required:"true"`
	// The name or ID of the template used as base template for the instance. This option is mutually exclusive with source_iso.
	SourceTemplate         string   `mapstructure:"source_template" required:"true"`
	TemporaryKeypairName   string   `mapstructure:"temporary_keypair_name"`
	UseLocalIPAddress      bool     `mapstructure:"use_local_ip_address"`
	UserData               string   `mapstructure:"user_data"`
	UserDataFile           string   `mapstructure:"user_data_file"`
	// The name or ID of the zone where the instance will be created.
	Zone                   string   `mapstructure:"zone" required:"true"`

	TemplateName            string `mapstructure:"template_name"`
	TemplateDisplayText     string `mapstructure:"template_display_text"`
	// The name or ID of the template OS for the new template that will be created.
	TemplateOS              string `mapstructure:"template_os" required:"true"`
	TemplateFeatured        bool   `mapstructure:"template_featured"`
	TemplatePublic          bool   `mapstructure:"template_public"`
	TemplatePasswordEnabled bool   `mapstructure:"template_password_enabled"`
	TemplateRequiresHVM     bool   `mapstructure:"template_requires_hvm"`
	TemplateScalable        bool   `mapstructure:"template_scalable"`
	TemplateTag             string `mapstructure:"template_tag"`

	ctx interpolate.Context
}

// NewConfig parses and validates the given config.
func NewConfig(raws ...interface{}) (*Config, error) {
	c := new(Config)
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"user_data",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	var errs *packer.MultiError

	// Set some defaults.
	if c.APIURL == "" {
		// Default to environment variable for api_url, if it exists
		c.APIURL = os.Getenv("CLOUDSTACK_API_URL")
	}

	if c.APIKey == "" {
		// Default to environment variable for api_key, if it exists
		c.APIKey = os.Getenv("CLOUDSTACK_API_KEY")
	}

	if c.SecretKey == "" {
		// Default to environment variable for secret_key, if it exists
		c.SecretKey = os.Getenv("CLOUDSTACK_SECRET_KEY")
	}

	if c.AsyncTimeout == 0 {
		c.AsyncTimeout = 30 * time.Minute
	}

	if len(c.CIDRList) == 0 {
		c.CIDRList = []string{"0.0.0.0/0"}
	}

	if c.InstanceName == "" {
		c.InstanceName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.TemplateName == "" {
		name, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Unable to parse template name: %s ", err))
		}

		c.TemplateName = name
	}

	if c.TemplateDisplayText == "" {
		c.TemplateDisplayText = c.TemplateName
	}

	// If we are not given an explicit keypair, ssh_password or ssh_private_key_file,
	// then create a temporary one, but only if the temporary_keypair_name has not
	// been provided.
	if c.Comm.SSHKeyPairName == "" && c.Comm.SSHTemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKeyFile == "" && c.Comm.SSHPassword == "" {
		c.Comm.SSHTemporaryKeyPairName = fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	}

	// Process required parameters.
	if c.APIURL == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a api_url must be specified"))
	}

	if c.APIKey == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a api_key must be specified"))
	}

	if c.SecretKey == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a secret_key must be specified"))
	}

	if c.Network == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a network must be specified"))
	}

	if c.CreateSecurityGroup && !c.Expunge {
		errs = packer.MultiErrorAppend(errs, errors.New("auto creating a temporary security group requires expunge"))
	}

	if c.ServiceOffering == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a service_offering must be specified"))
	}

	if c.SourceISO == "" && c.SourceTemplate == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("either source_iso or source_template must be specified"))
	}

	if c.SourceISO != "" && c.SourceTemplate != "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("only one of source_iso or source_template can be specified"))
	}

	if c.SourceISO != "" && c.DiskOffering == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a disk_offering must be specified when using source_iso"))
	}

	if c.SourceISO != "" && c.Hypervisor == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("a hypervisor must be specified when using source_iso"))
	}

	if c.TemplateOS == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a template_os must be specified"))
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("only one of user_data or user_data_file can be specified"))
	}

	if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("user_data_file not found: %s", c.UserDataFile))
		}
	}

	if c.Zone == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("a zone must be specified"))
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	// Check for errors and return if we have any.
	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return c, nil
}
