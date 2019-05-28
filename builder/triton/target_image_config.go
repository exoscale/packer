package triton

import (
	"fmt"

	"github.com/hashicorp/packer/template/interpolate"
)

// TargetImageConfig represents the configuration for the image to be created
// from the source machine.
type TargetImageConfig struct {
	// The name the finished image in Triton will be assigned. Maximum 512 characters but should in practice be much shorter (think between 5 and 20 characters). For example postgresql-95-server for an image used as a PostgreSQL 9.5 server.
	ImageName        string            `mapstructure:"image_name" required:"true"`
	// The version string for this image. Maximum 128 characters. Any string will do but a format of Major.Minor.Patch is strongly advised by Joyent. See Semantic Versioning for more information on the Major.Minor.Patch versioning format.
	ImageVersion     string            `mapstructure:"image_version" required:"true"`
	ImageDescription string            `mapstructure:"image_description"`
	ImageHomepage    string            `mapstructure:"image_homepage"`
	ImageEULA        string            `mapstructure:"image_eula_url"`
	ImageACL         []string          `mapstructure:"image_acls"`
	ImageTags        map[string]string `mapstructure:"image_tags"`
}

// Prepare performs basic validation on a TargetImageConfig struct.
func (c *TargetImageConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.ImageName == "" {
		errs = append(errs, fmt.Errorf("An image_name must be specified"))
	}

	if c.ImageVersion == "" {
		errs = append(errs, fmt.Errorf("An image_version must be specified"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
