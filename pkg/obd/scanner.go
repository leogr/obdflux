package obd

import (
	"strconv"
	"time"

	"github.com/rzetterberg/elmobd"
)

// A Scanner periodically requests data from an OBD device and sends resulting measurements through its output channel.
type Scanner struct {
	dev  *elmobd.Device
	cmds []elmobd.OBDCommand
	t    *time.Ticker
	ch   chan []Measurement
	stop chan struct{}
}

// NewScanner returns a new Scanner bound to the elmobd.Device dev.
// Scanner runs OBD commands cmds repeatedly with a period specified by the d argument
// and sends resulting []Measurement to the output a channel.
// Channel buffer size can be set using the capacity argument.
// Stop the scanner to release associated resources.
func NewScanner(dev *elmobd.Device, cmds []elmobd.OBDCommand, d time.Duration, capacity int) *Scanner {
	s := &Scanner{
		dev,
		cmds,
		time.NewTicker(d),
		make(chan []Measurement, capacity),
		make(chan struct{}, 1),
	}

	go s.run()

	return s
}

func (s *Scanner) run() {
	defer close(s.ch)
	defer close(s.stop)

	for {
		select {
		case <-s.t.C:
			res := make([]Measurement, len(s.cmds))
			var cmdRes elmobd.OBDCommand
			for i, cmd := range s.cmds {

				m := &res[i]

				m.ModeID = cmd.ModeID()
				m.ParameterID = byte(cmd.ParameterID())
				m.Key = cmd.Key()

				cmdRes, m.Err = s.dev.RunOBDCommand(cmd)
				m.Time = time.Now()

				if m.Err == nil {
					// (fixme): temp workaround to get a float from elmobd.OBDCommand
					f, err := strconv.ParseFloat(cmdRes.ValueAsLit(), 64)
					if err != nil {
						m.Value = cmdRes.ValueAsLit()
					} else {
						m.Value = f
					}
				}
			}
			s.ch <- res
		case <-s.stop:
			s.t.Stop()
			return
		}
	}
}

// C returns the output channel on which the []Measurement are delivered.
func (s *Scanner) C() <-chan []Measurement {
	return s.ch
}

// Stop turns off a scanner. After Stop, the channel output will be closed.
// Stop will panic if called on an alredy stopped scanner.
func (s *Scanner) Stop() {
	s.stop <- struct{}{}
}
