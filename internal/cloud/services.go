package cloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	supabaseURL = "https://xqmjdjjwprnqogokoejz.supabase.co"
	cliSecret   = "rukkie-cli-v1-xqmjdjjwprnqogokoejz"
)

// ServiceStatus holds the cloud status for a registered service.
type ServiceStatus struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Language    string  `json:"language"`
	Description string  `json:"description"`
	ActiveKeys  int     `json:"activeKeys"`
	LastUsedAt  *string `json:"lastUsedAt"`
}

type servicesResponse struct {
	Services []ServiceStatus `json:"services"`
	Error    string          `json:"error"`
}

// FetchServices calls the services-status Edge Function and returns all registered services.
func FetchServices() ([]ServiceStatus, error) {
	url := supabaseURL + "/functions/v1/services-status"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("x-rukkie-cli", cliSecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach RukkiePulse cloud: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result servicesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid response from cloud: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("cloud error: %s", result.Error)
	}

	return result.Services, nil
}

// ConnectionStatus returns the emoji dot, label, and last-seen text for a service.
func ConnectionStatus(lastUsedAt *string) (dot, label, lastSeen string) {
	if lastUsedAt == nil || *lastUsedAt == "" {
		return "⚫", "Never connected", ""
	}

	t, err := time.Parse(time.RFC3339, *lastUsedAt)
	if err != nil {
		return "⚫", "Never connected", ""
	}

	ago := time.Since(t)
	mins := ago.Minutes()

	switch {
	case mins < 5:
		return "🟢", "Live", fmt.Sprintf("last seen %dm ago", int(mins))
	case mins < 60:
		return "🟡", "Recent", fmt.Sprintf("last seen %dm ago", int(mins))
	default:
		return "🔴", "Inactive", fmt.Sprintf("last seen %s", t.Format("2006-01-02 15:04"))
	}
}
