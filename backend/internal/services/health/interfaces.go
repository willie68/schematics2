package health

import "errors"

// Check is implemented by components that participate in health checks.
type Check interface {
	// CheckName returns a unique healthcheck name.
	CheckName() string
	// Check performs a health check and returns status + optional error.
	Check() (bool, error)
}

// Registerer registers checks in the health system.
type Registerer interface {
	Register(check Check)
}

// Unregisterer unregisters checks from the health system.
type Unregisterer interface {
	Unregister(checkname string) bool
}

// Register registers a new health check.
func Register(r Registerer, check Check) error {
	if r == nil {
		return errors.New("health registerer is nil")
	}
	r.Register(check)
	return nil
}

// Unregister removes a health check by name.
func Unregister(u Unregisterer, checkname string) error {
	if u == nil {
		return errors.New("health unregisterer is nil")
	}
	if !u.Unregister(checkname) {
		return errors.New("check with name not found")
	}
	return nil
}
