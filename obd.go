package main

import (
	"strconv"
	"time"

	"github.com/rzetterberg/elmobd"
)

type OBDMeasurement struct {
	time   time.Time
	values map[string]interface{}
	errors map[string]string
}

func (m *OBDMeasurement) Time() time.Time {
	return m.time
}

func (m *OBDMeasurement) Values() map[string]interface{} {
	return m.values
}

func (m *OBDMeasurement) Errors() map[string]string {
	return m.errors
}

type OBDScanner struct {
	dev   *elmobd.Device
	cmds  []elmobd.OBDCommand
	sleep time.Duration
	ch    chan Measurement
	stop  chan struct{}
}

func NewOBDScanner(dev *elmobd.Device, cmds []elmobd.OBDCommand, sleep time.Duration, capacity int) *OBDScanner {
	return &OBDScanner{
		dev,
		cmds,
		sleep,
		make(chan Measurement, capacity),
		make(chan struct{}),
	}
}

func (o *OBDScanner) Measurements() <-chan Measurement {
	return o.ch
}

func (o *OBDScanner) Start() {
	for {

		select {
		case <-o.stop:
			return
		default:
		}

		values := map[string]interface{}{}
		errors := map[string]string{}

		for _, cmd := range o.cmds {
			cmdRes, cmdErr := o.dev.RunOBDCommand(cmd)

			if cmdErr != nil {
				errors["err_"+cmdRes.Key()] = cmdErr.Error()
			} else {
				f, err := strconv.ParseFloat(cmdRes.ValueAsLit(), 64)
				if err != nil {
					values[cmdRes.Key()] = cmdRes.ValueAsLit()
				} else {
					values[cmdRes.Key()] = f
				}
			}
		}

		o.ch <- &OBDMeasurement{
			time.Now(),
			values,
			errors,
		}

		time.Sleep(o.sleep)
	}
}

func (o *OBDScanner) Stop() {
	o.stop <- struct{}{}
}
