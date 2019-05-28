package common

import (
	"fmt"
	"time"

	"github.com/hashicorp/packer/template/interpolate"
)

type ShutdownConfig struct {
	// The command to use to gracefully shut down the machine once all the provisioning is done. By default this is an empty string, which tells Packer to just forcefully shut down the machine unless a shutdown command takes place inside script so this may safely be omitted. If one or more scripts require a reboot it is suggested to leave this blank since reboots may fail and specify the final shutdown command in your last script.
	ShutdownCommand    string `mapstructure:"shutdown_command" required:"false"`
	// The amount of time to wait after executing the shutdown_command for the virtual machine to actually shut down. If it doesn't shut down in this time, it is an error. By default, the timeout is 5m or five minutes.
	RawShutdownTimeout string `mapstructure:"shutdown_timeout" required:"false"`

	ShutdownTimeout time.Duration ``
}

func (c *ShutdownConfig) Prepare(ctx *interpolate.Context) []error {
	if c.RawShutdownTimeout == "" {
		c.RawShutdownTimeout = "5m"
	}

	var errs []error
	var err error
	c.ShutdownTimeout, err = time.ParseDuration(c.RawShutdownTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
	}

	return errs
}
