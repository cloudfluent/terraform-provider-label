package provider

import (
	"strings"
)

// LabelConfig holds workspace-level naming values sourced from environment variables.
type LabelConfig struct {
	Tenant      string
	Environment string
	Stage       string
	Workspace   string
	Namespace   string // optional, used in tags only
	Delimiter   string // default "-", override via LABEL_DELIMITER
}

// SplitWorkspace splits the workspace name by "-" into segments.
// e.g. "sales-api" → ["sales", "api"], "vpc" → ["vpc"], "" → []
func SplitWorkspace(workspace string) []string {
	if workspace == "" {
		return nil
	}
	return strings.Split(workspace, "-")
}

// GenerateID builds an identifier string from the label components.
// Order: {tenant}{d}{environment}{d}{resource_type}{d}{stage}{d}{qualifier}{d}{workspace}{d}{instance_key}
// Empty segments are skipped.
func GenerateID(cfg *LabelConfig, resourceType string, qualifier string, instanceKey string, delimiter string) string {
	if delimiter == "" {
		delimiter = cfg.Delimiter
	}

	ws := SplitWorkspace(cfg.Workspace)

	parts := []string{
		cfg.Tenant,
		cfg.Environment,
		resourceType,
		cfg.Stage,
	}

	if qualifier != "" {
		parts = append(parts, qualifier)
	}

	parts = append(parts, ws...)

	if instanceKey != "" {
		parts = append(parts, instanceKey)
	}

	return strings.Join(parts, delimiter)
}

// GenerateTags builds a tag map for the resource.
func GenerateTags(cfg *LabelConfig, resourceType string, qualifier string, instanceKey string, delimiter string) map[string]string {
	name := GenerateID(cfg, resourceType, qualifier, instanceKey, delimiter)

	ws := SplitWorkspace(cfg.Workspace)

	var attrParts []string
	if qualifier != "" {
		attrParts = append(attrParts, qualifier)
	}
	attrParts = append(attrParts, ws...)
	if instanceKey != "" {
		attrParts = append(attrParts, instanceKey)
	}

	attributes := strings.Join(attrParts, "-")

	tags := map[string]string{
		"Name":        name,
		"Tenant":      cfg.Tenant,
		"Environment": cfg.Environment,
		"Stage":       cfg.Stage,
	}

	if cfg.Namespace != "" {
		tags["Namespace"] = cfg.Namespace
	}

	if attributes != "" {
		tags["Attributes"] = attributes
	}

	return tags
}
