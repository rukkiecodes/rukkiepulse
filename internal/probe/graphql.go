package probe

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rukkiecodes/rukkiepulse/internal/config"
)

func CheckGraphQL(baseURL string, ep config.Endpoint) EndpointResult {
	query := ep.Query
	if query == "" {
		query = "{ __typename }"
	}

	payload := fmt.Sprintf(`{"query": %q}`, query)

	start := time.Now()
	resp, err := httpClient.Post(baseURL, "application/json", strings.NewReader(payload))
	latency := time.Since(start)

	if err != nil {
		return EndpointResult{
			Path: baseURL, Method: "POST", Kind: "GRAPHQL",
			Status: "fail", Latency: latency,
			Error: fmt.Sprintf("connection failed: %v", err),
		}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return EndpointResult{
			Path: baseURL, Method: "POST", Kind: "GRAPHQL",
			Status: "fail", Latency: latency, Code: resp.StatusCode,
			Error: "could not parse GraphQL response",
		}
	}

	status := "pass"
	errMsg := ""
	if ep.ExpectNoErrors {
		if errors, ok := result["errors"]; ok && errors != nil {
			status = "fail"
			errMsg = fmt.Sprintf("GraphQL errors: %v", errors)
		}
	}

	if resp.StatusCode != 200 {
		status = "fail"
		errMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return EndpointResult{
		Path:    baseURL,
		Method:  "POST",
		Kind:    "GRAPHQL",
		Status:  status,
		Code:    resp.StatusCode,
		Latency: latency,
		Error:   errMsg,
	}
}
