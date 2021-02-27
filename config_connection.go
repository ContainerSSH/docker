package docker

import (
	"github.com/containerssh/log"
)

func (c ConnectionConfig) Validate() error {
	if c.Host == "" {
		return log.NewMessage(EConfigError, "missing host")
	}
	return nil
}
