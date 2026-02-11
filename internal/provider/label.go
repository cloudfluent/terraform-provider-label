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

// ExtractDomain splits the workspace name and drops the first token.
// e.g. "data-sales-api" → ["sales", "api"], "vpc" → []
func ExtractDomain(workspace string) []string {
	parts := strings.Split(workspace, "-")
	if len(parts) <= 1 {
		return nil
	}
	return parts[1:]
}

// GenerateID builds an identifier string from the label components.
// Order: {tenant}{d}{environment}{d}{resource_type}{d}{stage}{d}{qualifier}{d}{domain}{d}{instance_key}
// Empty segments are skipped.
func GenerateID(cfg *LabelConfig, resourceType string, qualifier string, instanceKey string, delimiter string) string {
	if delimiter == "" {
		delimiter = cfg.Delimiter
	}

	domain := ExtractDomain(cfg.Workspace)

	parts := []string{
		cfg.Tenant,
		cfg.Environment,
		resourceType,
		cfg.Stage,
	}

	if qualifier != "" {
		parts = append(parts, qualifier)
	}

	parts = append(parts, domain...)

	if instanceKey != "" {
		parts = append(parts, instanceKey)
	}

	return strings.Join(parts, delimiter)
}

// GenerateTags builds a tag map for the resource.
func GenerateTags(cfg *LabelConfig, resourceType string, qualifier string, instanceKey string, delimiter string) map[string]string {
	name := GenerateID(cfg, resourceType, qualifier, instanceKey, delimiter)

	domain := ExtractDomain(cfg.Workspace)

	var attrParts []string
	if qualifier != "" {
		attrParts = append(attrParts, qualifier)
	}
	attrParts = append(attrParts, domain...)
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
