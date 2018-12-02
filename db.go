package main

import (
	"fmt"

	"github.com/leogr/obdflux/pkg/obd"

	"github.com/influxdata/influxdb/client/v2"
)

type DBClient struct {
	measurement string
	client      client.Client
	bpCfg       client.BatchPointsConfig
}

func NewDBClient(conf client.HTTPConfig, db string, measurement string) (*DBClient, error) {
	c, err := client.NewHTTPClient(conf)
	if err != nil {
		return nil, err
	}

	bpCfg := client.BatchPointsConfig{
		Database: db,
	}

	return &DBClient{
		measurement,
		c,
		bpCfg,
	}, nil
}

func (d *DBClient) Write(ms []obd.Measurement) error {
	bps, err := client.NewBatchPoints(d.bpCfg)
	if err != nil {
		return err
	}

	for _, m := range ms {
		tags := map[string]string{
			"mode": fmt.Sprintf("%02X", m.ModeID),
			"PID":  fmt.Sprintf("%02X", m.ParameterID),
		}
		fields := map[string]interface{}{}
		if m.Err != nil {
			tags["msg"] = m.Err.Error()
			fields["err"] = true
		} else {
			fields[m.Key] = m.Value
		}
		p, err := client.NewPoint(
			d.measurement,
			tags,
			fields,
			m.Time,
		)
		if err != nil {
			return err
		}
		bps.AddPoint(p)
	}

	return d.client.Write(bps)
}

func (d *DBClient) Close() error {
	return d.client.Close()
}
