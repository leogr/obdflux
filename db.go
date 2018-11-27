package main

import (
	"time"

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

func (d *DBClient) Write(tags map[string]string, fields map[string]interface{}, time time.Time) error {
	bps, err := client.NewBatchPoints(d.bpCfg)
	if err != nil {
		return err
	}

	p, err := client.NewPoint(d.measurement, tags, fields, time)
	if err != nil {
		return err
	}

	bps.AddPoint(p)
	return d.client.Write(bps)
}

func (d *DBClient) Close() error {
	return d.client.Close()
}
