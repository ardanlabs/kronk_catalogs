//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

type Catalog struct {
	Catalog string  `yaml:"catalog"`
	Models  []Model `yaml:"models"`
}

type Model struct {
	ID         string `yaml:"id"`
	Category   string `yaml:"category"`
	OwnedBy    string `yaml:"owned_by"`
	GatedModel bool   `yaml:"gated_model"`
	WebPage    string `yaml:"web_page"`
	Files      struct {
		Models []File `yaml:"models"`
		Proj   *File  `yaml:"proj,omitempty"`
	} `yaml:"files"`
	Capabilities struct {
		Streaming bool `yaml:"streaming"`
		Reasoning bool `yaml:"reasoning"`
		Audio     bool `yaml:"audio"`
		Video     bool `yaml:"video"`
		Tooling   bool `yaml:"tooling"`
		Images    bool `yaml:"images"`
		Embedding bool `yaml:"embedding"`
		Rerank    bool `yaml:"rerank"`
	} `yaml:"capabilities"`
	Metadata struct {
		Created     time.Time `yaml:"created"`
		Description string    `yaml:"description"`
	} `yaml:"metadata"`
	Config struct {
		ContextWindow int `yaml:"context-window"`
	} `yaml:"config"`
}

type File struct {
	URL  string `yaml:"url"`
	Size string `yaml:"size"`
}

func main() {
	catalogsDir := "catalogs"
	outputFile := "CATALOG.md"

	files, err := filepath.Glob(filepath.Join(catalogsDir, "*.yaml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var allModels []Model
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: read %q: %v\n", file, err)
			os.Exit(1)
		}

		var catalog Catalog
		if err := yaml.Unmarshal(data, &catalog); err != nil {
			fmt.Fprintf(os.Stderr, "error: parse %q: %v\n", file, err)
			os.Exit(1)
		}

		allModels = append(allModels, catalog.Models...)
	}

	sort.Slice(allModels, func(i, j int) bool {
		return strings.ToLower(allModels[i].ID) < strings.ToLower(allModels[j].ID)
	})

	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Fprintln(f, "# Model Catalog")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "| ID | Category | Owner | Size | Context | Created | Description | a<br>u<br>d<br>i<br>o | g<br>a<br>t<br>e<br>d | r<br>e<br>a<br>s<br>o<br>n | s<br>t<br>r<br>e<br>a<br>m | t<br>o<br>o<br>l | v<br>i<br>d<br>e<br>o |")
	fmt.Fprintln(f, "|---|---|---|---|---|---|---|:---:|:---:|:---:|:---:|:---:|:---:|")

	for _, model := range allModels {
		id := fmt.Sprintf("[%s](%s)", model.ID, model.WebPage)
		category := model.Category
		owner := model.OwnedBy
		size := calculateTotalSize(model.Files)
		context := formatContextWindow(model.Config.ContextWindow)
		created := model.Metadata.Created.Format("2006-01-02")
		description := formatDescription(model.Metadata.Description)

		audio := boolToMark(model.Capabilities.Audio)
		gated := boolToMark(model.GatedModel)
		reason := boolToMark(model.Capabilities.Reasoning)
		stream := boolToMark(model.Capabilities.Streaming)
		tool := boolToMark(model.Capabilities.Tooling)
		video := boolToMark(model.Capabilities.Video)

		fmt.Fprintf(f, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			id, category, owner, size, context, created, description, audio, gated, reason, stream, tool, video)
	}

	fmt.Printf("Generated %s with %d models\n", outputFile, len(allModels))
}

func calculateTotalSize(files struct {
	Models []File `yaml:"models"`
	Proj   *File  `yaml:"proj,omitempty"`
}) string {
	var totalBytes float64

	for _, model := range files.Models {
		totalBytes += parseSizeToBytes(model.Size)
	}

	if files.Proj != nil {
		totalBytes += parseSizeToBytes(files.Proj.Size)
	}

	if totalBytes >= GB {
		return fmt.Sprintf("%.1f GB", totalBytes/GB)
	}

	return fmt.Sprintf("%.0f MB", totalBytes/MB)
}

func parseSizeToBytes(size string) float64 {
	size = strings.TrimSpace(size)
	var value float64
	var unit string

	fmt.Sscanf(size, "%f %s", &value, &unit)

	unit = strings.ToLower(unit)
	switch unit {
	case "GB", "gb":
		return value * GB
	case "MB", "mb":
		return value * MB
	case "kib", "kb":
		return value * KB
	default:
		return value
	}
}

func boolToMark(b bool) string {
	if b {
		return "✓"
	}
	return ""
}

func formatDescription(desc string) string {
	if desc == "" {
		return ""
	}

	desc = strings.TrimSpace(desc)
	desc = strings.ReplaceAll(desc, "\n", " ")
	return fmt.Sprintf("<details><summary>Show</summary>%s</details>", desc)
}

func formatContextWindow(size int) string {
	if size == 0 {
		return "-"
	}

	if size >= KB {
		return fmt.Sprintf("%dK", size/KB)
	}

	return fmt.Sprintf("%d", size)
}
