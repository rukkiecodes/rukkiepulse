package engine

import (
	"sync"

	"github.com/rukkiecodes/rukkiepulse/internal/config"
	"github.com/rukkiecodes/rukkiepulse/internal/health"
	"github.com/rukkiecodes/rukkiepulse/internal/probe"
)

func Run(services []config.Service) []ServiceResult {
	results := make(chan ServiceResult, len(services))
	var wg sync.WaitGroup

	for _, svc := range services {
		wg.Add(1)
		go func(s config.Service) {
			defer wg.Done()
			results <- checkService(s)
		}(svc)
	}

	wg.Wait()
	close(results)

	// collect into a map then restore config order
	byName := make(map[string]ServiceResult, len(services))
	for r := range results {
		byName[r.Name] = r
	}

	ordered := make([]ServiceResult, 0, len(services))
	for _, svc := range services {
		if r, ok := byName[svc.Name]; ok {
			ordered = append(ordered, r)
		}
	}
	return ordered
}

func checkService(svc config.Service) ServiceResult {
	result := ServiceResult{
		Name: svc.Name,
		URL:  svc.URL,
		Type: svc.Type,
	}

	// health check + endpoint probes run concurrently
	type healthOut struct{ r health.Result }
	type probeOut struct{ r probe.EndpointResult }

	healthCh := make(chan healthOut, 1)
	probeCh := make(chan probeOut, len(svc.Endpoints))

	go func() {
		healthCh <- healthOut{health.Check(svc)}
	}()

	var wg sync.WaitGroup
	for _, ep := range svc.Endpoints {
		wg.Add(1)
		go func(e config.Endpoint) {
			defer wg.Done()
			if svc.Type == "GRAPHQL" && e.Query != "" {
				probeCh <- probeOut{probe.CheckGraphQL(svc.URL, e)}
			} else {
				probeCh <- probeOut{probe.CheckREST(svc.URL, e)}
			}
		}(ep)
	}

	wg.Wait()
	close(probeCh)

	result.Health = (<-healthCh).r
	for p := range probeCh {
		result.Endpoints = append(result.Endpoints, p.r)
	}

	return result
}
