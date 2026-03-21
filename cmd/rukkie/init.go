package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rukkiecodes/rukkiepulse/internal/auth"
	"github.com/rukkiecodes/rukkiepulse/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type rukkieYAML struct {
	Service struct {
		Name     string `yaml:"name"`
		Language string `yaml:"language"`
		APIKey   string `yaml:"apiKey"`
	} `yaml:"service"`
	Observability struct {
		Jaeger struct {
			URL string `yaml:"url"`
		} `yaml:"jaeger"`
		Collector string `yaml:"collector"`
	} `yaml:"observability"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create rukkie.yaml and print the integration snippet for this project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.RequireAuth(); err != nil {
			return err
		}

		cwd, _ := os.Getwd()
		lang := detectLanguage(cwd)
		name := filepath.Base(cwd)

		yamlPath := filepath.Join(cwd, "rukkie.yaml")

		// Don't overwrite if already exists
		if _, err := os.Stat(yamlPath); err == nil {
			output.PrintError("rukkie.yaml already exists in this directory.")
			return nil
		}

		cfg := rukkieYAML{}
		cfg.Service.Name = name
		cfg.Service.Language = lang
		cfg.Service.APIKey = "YOUR_API_KEY"
		cfg.Observability.Jaeger.URL = "http://localhost:16686"
		cfg.Observability.Collector = "http://localhost:4317"

		data, err := yaml.Marshal(&cfg)
		if err != nil {
			return err
		}
		if err := os.WriteFile(yamlPath, data, 0644); err != nil {
			return err
		}

		fmt.Printf("\n  ✅  Created rukkie.yaml for \"%s\" (%s)\n\n", name, lang)
		fmt.Println("  Replace YOUR_API_KEY with a key from the dashboard:")
		fmt.Println("  https://rukkiepulse-dashboard.netlify.app\n")
		printInitSnippet(name, lang)
		return nil
	},
}

func detectLanguage(dir string) string {
	checks := map[string]string{
		"package.json":   "node",
		"pyproject.toml": "python",
		"requirements.txt": "python",
		"go.mod":         "go",
	}
	for file, lang := range checks {
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			return lang
		}
	}
	return "other"
}

func printInitSnippet(name, lang string) {
	switch lang {
	case "node":
		fmt.Println("  ── ESM ────────────────────────────────────────────────")
		fmt.Printf("  import { initRukkie } from 'rukkie-agent'\n\n")
		fmt.Printf("  initRukkie({\n    serviceName: '%s',\n    apiKey: 'YOUR_API_KEY',\n  })\n\n", name)
		fmt.Println("  ── CommonJS ───────────────────────────────────────────")
		fmt.Printf("  const { initRukkie } = require('rukkie-agent')\n\n")
		fmt.Printf("  initRukkie({\n    serviceName: '%s',\n    apiKey: 'YOUR_API_KEY',\n  })\n\n", name)
		fmt.Println("  Install:  npm install rukkie-agent")
	case "python":
		fmt.Println("  ── Python ─────────────────────────────────────────────")
		fmt.Printf("  from rukkie_agent import init_rukkie\n\n")
		fmt.Printf("  init_rukkie(\n      service_name=\"%s\",\n      api_key=\"YOUR_API_KEY\",\n  )\n\n", name)
		fmt.Println("  Install:  pip install rukkie-agent")
	case "go":
		fmt.Println("  ── Go ─────────────────────────────────────────────────")
		fmt.Println("  Go agent coming soon. Use rukkie scan to monitor from the CLI.")
	default:
		fmt.Println("  ── REST ───────────────────────────────────────────────")
		fmt.Println("  POST /api/v1/heartbeat")
		fmt.Println("  Authorization: Bearer YOUR_API_KEY")
	}
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(initCmd)
}
