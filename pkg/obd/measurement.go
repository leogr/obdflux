package obd

import "time"

// A Measurement holds the resulting value of an OBD measurement at a specific point in time.
type Measurement struct {
	ModeID      byte
	ParameterID byte
	Key         string
	Time        time.Time
	Value       interface{}
	Err         error
}
