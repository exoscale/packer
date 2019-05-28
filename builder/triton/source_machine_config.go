package triton

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

// SourceMachineConfig represents the configuration to run a machine using
// the SDC API in order for provisioning to take place.
type SourceMachineConfig struct {
	MachineName            string             `mapstructure:"source_machine_name"`
	// The Triton package to use while building the image. Does not affect (and does not have to be the same) as the package which will be used for a VM instance running this image. On the Joyent public cloud this could for example be g3-standard-0.5-smartos.
	MachinePackage         string             `mapstructure:"source_machine_package" required:"true"`
	// The UUID of the image to base the new image on. Triton supports multiple types of images, called 'brands' in Triton / Joyent lingo, for contains and VM's. See the chapter Containers and virtual machines in the Joyent Triton documentation for detailed information. The following brands are currently supported by this builder:joyent andkvm. The choice of base image automatically decides the brand. On the Joyent public cloud a valid source_machine_image could for example be 70e3ae72-96b6-11e6-9056-9737fd4d0764 for version 16.3.1 of the 64bit SmartOS base image (a 'joyent' brand image). source_machine_image_filter can be used to populate this UUID.
	MachineImage           string             `mapstructure:"source_machine_image" required:"true"`
	MachineNetworks        []string           `mapstructure:"source_machine_networks"`
	MachineMetadata        map[string]string  `mapstructure:"source_machine_metadata"`
	MachineTags            map[string]string  `mapstructure:"source_machine_tags"`
	MachineFirewallEnabled bool               `mapstructure:"source_machine_firewall_enabled"`
	MachineImageFilters    MachineImageFilter `mapstructure:"source_machine_image_filter"`
}

type MachineImageFilter struct {
	MostRecent bool `mapstructure:"most_recent"`
	Name       string
	OS         string
	Version    string
	Public     bool
	State      string
	Owner      string
	Type       string
}

func (m *MachineImageFilter) Empty() bool {
	return m.Name == "" && m.OS == "" && m.Version == "" && m.State == "" && m.Owner == "" && m.Type == ""
}

// Prepare performs basic validation on a SourceMachineConfig struct.
func (c *SourceMachineConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.MachinePackage == "" {
		errs = append(errs, fmt.Errorf("A source_machine_package must be specified"))
	}

	if c.MachineImage != "" && c.MachineImageFilters.Name != "" {
		errs = append(errs, fmt.Errorf("You cannot specify a Machine Image and also Machine Name filter"))
	}

	if c.MachineNetworks == nil {
		c.MachineNetworks = []string{}
	}

	if c.MachineMetadata == nil {
		c.MachineMetadata = make(map[string]string)
	}

	if c.MachineTags == nil {
		c.MachineTags = make(map[string]string)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
