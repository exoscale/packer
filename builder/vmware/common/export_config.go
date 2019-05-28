package common

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

type ExportConfig struct {
	// Either "ovf", "ova" or "vmx", this specifies the output format of the exported virtual machine. This defaults to "ovf". Before using this option, you need to install ovftool.
	Format         string   `mapstructure:"format" required:"false"`
	OVFToolOptions []string `mapstructure:"ovftool_options"`
	// Defaults to false. When enabled, Packer will not export the VM. Useful if the build output is not the resultant image, but created inside the VM.
	SkipExport     bool     `mapstructure:"skip_export" required:"false"`
	// Set this to true if you would like to keep the VM registered with the remote ESXi server. This is convenient if you use packer to provision VMs on ESXi and don't want to use ovftool to deploy the resulting artifact (VMX or OVA or whatever you used as format). Defaults to false.
	KeepRegistered bool     `mapstructure:"keep_registered" required:"false"`
	// VMware-created disks are defragmented and compacted at the end of the build process using vmware-vdiskmanager. In certain rare cases, this might actually end up making the resulting disks slightly larger. If you find this to be the case, you can disable compaction using this configuration value. Defaults to false.
	SkipCompaction bool     `mapstructure:"skip_compaction" required:"false"`
}

func (c *ExportConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if c.Format != "" {
		if !(c.Format == "ova" || c.Format == "ovf" || c.Format == "vmx") {
			errs = append(
				errs, fmt.Errorf("format must be one of ova, ovf, or vmx"))
		}
	}
	return errs
}
