package engine

import (
	"github.com/rukkiecodes/rukkiepulse/internal/health"
	"github.com/rukkiecodes/rukkiepulse/internal/probe"
)

type ServiceResult struct {
	Name      string
	URL       string
	Type      string
	Health    health.Result
	Endpoints []probe.EndpointResult
}

func (r ServiceResult) PassingEndpoints() (pass, total int) {
	total = len(r.Endpoints)
	for _, ep := range r.Endpoints {
		if ep.Status == "pass" {
			pass++
		}
	}
	return
}
