package docker

import (
	"fmt"
)

func (c ConnectionConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("missing host")
	}
	return nil
}
