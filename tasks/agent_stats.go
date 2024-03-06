package tasks

import (
	"context"
	"log/slog"
	"time"

	"github.com/bwiggs/spacetraders-go/api"
	"github.com/go-faster/errors"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func LogAgentMetrics(client *api.Client) error {
	client.GetAgent(context.TODO(), api.GetAgentParams{})
	dat, err := client.GetMyAgent(context.TODO())
	if err != nil {
		slog.Error(errors.Wrap(err, "failed to fetch my agent call").Error())
		return err
	}

	err = postInflux("agent", "credits", dat.Data.Credits)
	if err != nil {
		return err
	}
	err = postInflux("agent", "ships", dat.Data.ShipCount)
	if err != nil {
		return err
	}
	return nil
}

// TODO: move this into a metrics or monitoring package
func postInflux(measurement string, field string, value any) error {
	slog.Debug("posting to influx", "measurement", measurement, "field", field, "value", value)
	client := influxdb2.NewClient("http://localhost:53086", "9zL3uLzosPFVKNLWHG2so1XQ1990MYSh69VUAFe_WyB5XRW4fwrm_eHMoxdt36qiAz1PwFGtLhX__7ofpVXdKA==")
	writeAPI := client.WriteAPIBlocking("spacetraders", "spacetraders")

	fields := make(map[string]interface{})
	fields[field] = value

	p := influxdb2.NewPoint(measurement, nil, fields, time.Now())

	// write point immediately
	err := writeAPI.WritePoint(context.Background(), p)
	if err != nil {
		return errors.Wrap(err, "failed to write point to influxdb")
	}

	return nil
}
