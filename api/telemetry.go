package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/percona/percona-everest-backend/cmd/config"
)

const (
	telemetryProductFamily = "PRODUCT_FAMILY_EVEREST"

	// delay the initial metrics to prevent flooding in case of many restarts.
	initialMetricsDelay = 5 * time.Minute
)

// Telemetry is the struct for telemetry reports.
type Telemetry struct {
	Reports []Report `json:"reports"`
}

// Report is a struct for a single telemetry report.
type Report struct {
	ID            string    `json:"id"`
	CreateTime    time.Time `json:"createTime"`
	InstanceID    string    `json:"instanceId"`
	ProductFamily string    `json:"productFamily"`
	Metrics       []Metric  `json:"metrics"`
}

// Metric represents key-value metrics.
type Metric struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (e *EverestServer) report(ctx context.Context, baseURL string, data Telemetry) error {
	b, err := json.Marshal(data)
	if err != nil {
		e.l.Error(errors.Join(err, errors.New("failed to marshal the telemetry report")))
		return err
	}

	url := fmt.Sprintf("%s/v1/telemetry/GenericReport", baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		e.l.Error(errors.Join(err, errors.New("failed to create http request")))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.l.Error(errors.Join(err, errors.New("failed to send telemetry request")))
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		e.l.Info("Telemetry service responded with http status ", resp.StatusCode)
	}
	return nil
}

// RunTelemetryJob runs background job for collecting telemetry.
func (e *EverestServer) RunTelemetryJob(ctx context.Context, c *config.EverestConfig) {
	e.l.Debug("Starting background jobs runner.")

	interval, err := time.ParseDuration(c.TelemetryInterval)
	if err != nil {
		e.l.Error(errors.Join(err, errors.New("could not parse telemetry interval")))
		return
	}

	timer := time.NewTimer(initialMetricsDelay)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			timer.Reset(interval)
			err = e.collectMetrics(ctx, c.TelemetryURL)
			if err != nil {
				e.l.Error(errors.Join(err, errors.New("failed to collect telemetry data")))
			}
		}
	}
}

func (e *EverestServer) collectMetrics(ctx context.Context, url string) error {
	// FIXME
	//	everestID, err := e.storage.GetEverestID(ctx)
	//	if err != nil {
	//		e.l.Error(errors.Join(err, errors.New("failed to get Everest settings")))
	//		return err
	//	}
	//
	//ks, err := e.storage.ListKubernetesClusters(ctx)
	//if err != nil {
	//	e.l.Error(errors.Join(err, errors.New("could not list Kubernetes clusters")))
	//	return err
	//}
	//if len(ks) == 0 {
	//	return nil
	//}
	//// FIXME: Revisit it once multi k8s support will be enabled
	//_, kubeClient, _, err := e.initKubeClient(ctx, ks[0].ID)
	//if err != nil {
	//	e.l.Error(errors.Join(err, errors.New("could not init kube client for config")))
	//	return err
	//}

	clusters, err := e.kubeClient.ListDatabaseClusters(ctx)
	if err != nil {
		e.l.Error(errors.Join(err, errors.New("failed to list database clusters")))
		return err
	}

	types := make(map[string]int, 3)
	for _, cl := range clusters.Items {
		types[string(cl.Spec.Engine.Type)]++
	}

	// key - the engine type, value - the amount of db clusters of that type
	metrics := make([]Metric, 0, 3)
	for key, val := range types {
		metrics = append(metrics, Metric{key, strconv.Itoa(val)})
	}

	report := Telemetry{
		[]Report{
			{
				ID:            uuid.NewString(),
				CreateTime:    time.Now(),
				InstanceID:    "everestID", // FIXME
				ProductFamily: telemetryProductFamily,
				Metrics:       metrics,
			},
		},
	}

	return e.report(ctx, url, report)
}
