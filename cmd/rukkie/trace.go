package main

import (
	"fmt"

	"github.com/rukkiecodes/rukkiepulse/internal/auth"
	"github.com/rukkiecodes/rukkiepulse/internal/config"
	"github.com/rukkiecodes/rukkiepulse/internal/output"
	"github.com/rukkiecodes/rukkiepulse/internal/trace"
	"github.com/spf13/cobra"
)

var traceCmd = &cobra.Command{
	Use:   "trace <service> [endpoint]",
	Short: "Fetch and display the latest trace for a service",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runTrace,
}

var traceIDFlag string
var traceLimitFlag int
var traceFlameFlag bool

func init() {
	traceCmd.Flags().StringVar(&traceIDFlag, "trace-id", "", "fetch a specific trace by ID")
	traceCmd.Flags().IntVar(&traceLimitFlag, "last", 1, "number of recent traces to show")
	traceCmd.Flags().BoolVar(&traceFlameFlag, "flame", false, "render as a flame graph instead of waterfall")
	rootCmd.AddCommand(traceCmd)
}

func runTrace(cmd *cobra.Command, args []string) error {
	if err := auth.RequireAuth(); err != nil {
		output.PrintError(err.Error())
		return nil
	}

	cfg, err := config.Load("rukkie.yaml")
	if err != nil {
		output.PrintError(err.Error())
		return nil
	}

	jaegerURL := cfg.Observability.Jaeger.URL
	if jaegerURL == "" {
		output.PrintError("Jaeger is not configured — add observability.jaeger.url to rukkie.yaml")
		return nil
	}

	serviceName := args[0]
	operation := ""
	if len(args) > 1 {
		operation = args[1]
	}

	client := trace.NewClient(jaegerURL)

	if traceIDFlag != "" {
		t, err := client.FetchTrace(traceIDFlag)
		if err != nil {
			output.PrintError(fmt.Sprintf("failed to fetch trace: %v", err))
			return nil
		}
		printTraceHeader(serviceName, operation)
		renderTrace(t)
		return nil
	}

	limit := traceLimitFlag
	if limit < 1 {
		limit = 1
	}

	traces, err := client.FetchTraces(serviceName, operation, limit)
	if err != nil {
		output.PrintError(fmt.Sprintf("failed to fetch traces from Jaeger: %v", err))
		return nil
	}
	if len(traces) == 0 {
		fmt.Printf("\n  No traces found for %q", serviceName)
		if operation != "" {
			fmt.Printf(" / %q", operation)
		}
		fmt.Println()
		return nil
	}

	printTraceHeader(serviceName, operation)
	for i, t := range traces {
		if len(traces) > 1 {
			fmt.Printf("── Trace %d of %d ──────────────────────────\n", i+1, len(traces))
		}
		renderTrace(&t)
	}

	return nil
}

func renderTrace(t *trace.Trace) {
	if traceFlameFlag {
		trace.RenderFlame(t)
	} else {
		trace.Render(t)
	}
}

func printTraceHeader(service, operation string) {
	fmt.Printf("\n\033[1mService:\033[0m %s", service)
	if operation != "" {
		fmt.Printf("  \033[1mEndpoint:\033[0m %s", operation)
	}
	fmt.Println()
}
