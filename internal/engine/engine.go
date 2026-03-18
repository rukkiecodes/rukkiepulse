package engine

import (
	"sync"

	"github.com/rukkiecodes/rukkiepulse/internal/config"
	"github.com/rukkiecodes/rukkiepulse/internal/health"
)

func Run(services []config.Service) []health.Result {
	results := make(chan health.Result, len(services))
	var wg sync.WaitGroup

	for _, svc := range services {
		wg.Add(1)
		go func(s config.Service) {
			defer wg.Done()
			results <- health.Check(s)
		}(svc)
	}

	wg.Wait()
	close(results)

	out := make([]health.Result, 0, len(services))
	for r := range results {
		out = append(out, r)
	}

	// restore original order from config
	ordered := make([]health.Result, 0, len(services))
	for _, svc := range services {
		for _, r := range out {
			if r.Service == svc.Name {
				ordered = append(ordered, r)
				break
			}
		}
	}

	return ordered
}
