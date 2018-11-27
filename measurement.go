package main

import (
	"time"
)

// Measurement holds values and errors of multiples
// OBD commands read at a specific point in time
type Measurement interface {
	Time() time.Time
	Values() map[string]interface{}
	Errors() map[string]string
}
