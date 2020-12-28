package docker

import (
	"errors"
	"time"
)

func parseRawDuration(rawValue interface{}, d *time.Duration) error {
	var err error
	switch value := rawValue.(type) {
	case int32:
		*d = time.Duration(value)
	case int64:
		*d = time.Duration(value)
	case int:
		*d = time.Duration(value)
	case float32:
		*d = time.Duration(value)
	case float64:
		*d = time.Duration(value)
	case string:
		if *d, err = time.ParseDuration(value); err != nil {
			return err
		}
	default:
		return errors.New("invalid duration")
	}
	return nil
}
