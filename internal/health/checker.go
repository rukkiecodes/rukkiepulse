package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rukkiecodes/rukkiepulse/internal/config"
)

type Response struct {
	Status       string            `json:"status"`
	Service      string            `json:"service"`
	Dependencies map[string]string `json:"dependencies"`
}

type Result struct {
	Service      string
	Status       string // "ok" | "degraded" | "down"
	Latency      time.Duration
	Error        string
	Dependencies map[string]string
}

var httpClient = &http.Client{Timeout: 5 * time.Second}

func Check(svc config.Service) Result {
	url := svc.URL + "/__rukkie/health"

	start := time.Now()
	resp, err := httpClient.Get(url)
	latency := time.Since(start)

	if err != nil {
		return Result{
			Service: svc.Name,
			Status:  "down",
			Latency: latency,
			Error:   fmt.Sprintf("connection failed: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{
			Service: svc.Name,
			Status:  "down",
			Latency: latency,
			Error:   fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	var health Response
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return Result{
			Service: svc.Name,
			Status:  "degraded",
			Latency: latency,
			Error:   "could not parse health response",
		}
	}

	status := "ok"
	if health.Status != "ok" {
		status = "degraded"
	}

	return Result{
		Service:      svc.Name,
		Status:       status,
		Latency:      latency,
		Dependencies: health.Dependencies,
	}
}
