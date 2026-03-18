package probe

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rukkiecodes/rukkiepulse/internal/config"
)

var httpClient = &http.Client{Timeout: 5 * time.Second}

type EndpointResult struct {
	Path    string
	Method  string
	Kind    string // "REST" | "GRAPHQL"
	Status  string // "pass" | "fail"
	Code    int
	Latency time.Duration
	Error   string
}

func CheckREST(baseURL string, ep config.Endpoint) EndpointResult {
	method := ep.Method
	if method == "" {
		method = "GET"
	}
	expectStatus := ep.ExpectStatus
	if expectStatus == 0 {
		expectStatus = 200
	}

	url := baseURL + ep.Path

	var bodyReader io.Reader
	if ep.Body != "" {
		bodyReader = strings.NewReader(ep.Body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return EndpointResult{
			Path: ep.Path, Method: method, Kind: "REST",
			Status: "fail", Error: fmt.Sprintf("bad request: %v", err),
		}
	}
	if ep.Body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := httpClient.Do(req)
	latency := time.Since(start)

	if err != nil {
		return EndpointResult{
			Path: ep.Path, Method: method, Kind: "REST",
			Status: "fail", Latency: latency,
			Error: fmt.Sprintf("connection failed: %v", err),
		}
	}
	defer resp.Body.Close()

	status := "pass"
	errMsg := ""
	if resp.StatusCode != expectStatus {
		status = "fail"
		errMsg = fmt.Sprintf("expected %d, got %d", expectStatus, resp.StatusCode)
	}

	return EndpointResult{
		Path:    ep.Path,
		Method:  method,
		Kind:    "REST",
		Status:  status,
		Code:    resp.StatusCode,
		Latency: latency,
		Error:   errMsg,
	}
}
