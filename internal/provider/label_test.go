package provider

import (
	"testing"
)

func TestSplitWorkspace(t *testing.T) {
	tests := []struct {
		workspace string
		want      []string
	}{
		{"sales-api", []string{"sales", "api"}},
		{"hr-web", []string{"hr", "web"}},
		{"vpc", []string{"vpc"}},
		{"eks-v1_34", []string{"eks", "v1_34"}},
		{"rds-postgres", []string{"rds", "postgres"}},
		{"sales-api-orders", []string{"sales", "api", "orders"}},
		{"", nil},
	}

	for _, tt := range tests {
		t.Run(tt.workspace, func(t *testing.T) {
			got := SplitWorkspace(tt.workspace)
			if len(got) != len(tt.want) {
				t.Fatalf("SplitWorkspace(%q) = %v, want %v", tt.workspace, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("SplitWorkspace(%q)[%d] = %q, want %q", tt.workspace, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	cfg := &LabelConfig{
		Tenant:      "dpl",
		Environment: "ane2",
		Stage:       "dev",
		Workspace:   "sales-api",
		Namespace:   "acme",
		Delimiter:   "-",
	}

	tests := []struct {
		name         string
		resourceType string
		qualifier    string
		instanceKey  string
		delimiter    string
		want         string
	}{
		{
			name:         "simple resource type",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev-sales-api",
		},
		{
			name:         "with qualifier",
			resourceType: "sg",
			qualifier:    "emr",
			want:         "dpl-ane2-sg-dev-emr-sales-api",
		},
		{
			name:         "with qualifier and instance key",
			resourceType: "role",
			qualifier:    "emr",
			instanceKey:  "shared-pii-etl",
			want:         "dpl-ane2-role-dev-emr-sales-api-shared-pii-etl",
		},
		{
			name:         "empty qualifier with instance key",
			resourceType: "emr",
			instanceKey:  "shared-pii",
			want:         "dpl-ane2-emr-dev-sales-api-shared-pii",
		},
		{
			name:         "custom delimiter",
			resourceType: "db",
			qualifier:    "refined",
			delimiter:    "_",
			want:         "dpl_ane2_db_dev_refined_sales_api",
		},
		{
			name:         "no qualifier",
			resourceType: "vpc",
			want:         "dpl-ane2-vpc-dev-sales-api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateID(cfg, tt.resourceType, tt.qualifier, tt.instanceKey, tt.delimiter)
			if got != tt.want {
				t.Errorf("GenerateID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateID_DifferentWorkspaces(t *testing.T) {
	tests := []struct {
		name         string
		workspace    string
		resourceType string
		want         string
	}{
		{
			name:         "single-segment workspace",
			workspace:    "vpc",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev-vpc",
		},
		{
			name:         "multi-segment workspace",
			workspace:    "sales-api-orders",
			resourceType: "task",
			want:         "dpl-ane2-task-dev-sales-api-orders",
		},
		{
			name:         "workspace with underscore",
			workspace:    "eks-v1_34",
			resourceType: "node",
			want:         "dpl-ane2-node-dev-eks-v1_34",
		},
		{
			name:         "empty workspace",
			workspace:    "",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &LabelConfig{
				Tenant:      "dpl",
				Environment: "ane2",
				Stage:       "dev",
				Workspace:   tt.workspace,
				Delimiter:   "-",
			}
			got := GenerateID(cfg, tt.resourceType, "", "", "")
			if got != tt.want {
				t.Errorf("GenerateID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateTags(t *testing.T) {
	cfg := &LabelConfig{
		Tenant:      "dpl",
		Environment: "ane2",
		Stage:       "dev",
		Workspace:   "sales-api",
		Namespace:   "acme",
		Delimiter:   "-",
	}

	tags := GenerateTags(cfg, "sg", "emr", "", "")

	expected := map[string]string{
		"Name":        "dpl-ane2-sg-dev-emr-sales-api",
		"Namespace":   "acme",
		"Tenant":      "dpl",
		"Environment": "ane2",
		"Stage":       "dev",
		"Attributes":  "emr-sales-api",
	}

	for k, v := range expected {
		if tags[k] != v {
			t.Errorf("tags[%q] = %q, want %q", k, tags[k], v)
		}
	}

	if len(tags) != len(expected) {
		t.Errorf("tags has %d keys, want %d", len(tags), len(expected))
	}
}

func TestGenerateTags_NoNamespace(t *testing.T) {
	cfg := &LabelConfig{
		Tenant:      "dpl",
		Environment: "ane2",
		Stage:       "dev",
		Workspace:   "sales-api",
		Delimiter:   "-",
	}

	tags := GenerateTags(cfg, "sg", "", "", "")

	if _, ok := tags["Namespace"]; ok {
		t.Error("Namespace should not be present when namespace is empty")
	}
}

func TestGenerateTags_WithInstanceKey(t *testing.T) {
	cfg := &LabelConfig{
		Tenant:      "dpl",
		Environment: "ane2",
		Stage:       "dev",
		Workspace:   "sales-api",
		Namespace:   "acme",
		Delimiter:   "-",
	}

	tags := GenerateTags(cfg, "role", "emr", "shared-pii-etl", "")

	if tags["Name"] != "dpl-ane2-role-dev-emr-sales-api-shared-pii-etl" {
		t.Errorf("Name = %q", tags["Name"])
	}
	if tags["Attributes"] != "emr-sales-api-shared-pii-etl" {
		t.Errorf("Attributes = %q", tags["Attributes"])
	}
}

func TestGenerateID_ConfigDelimiter(t *testing.T) {
	cfg := &LabelConfig{
		Tenant:      "dpl",
		Environment: "ane2",
		Stage:       "dev",
		Workspace:   "sales-api",
		Delimiter:   "_",
	}

	t.Run("config delimiter used as default", func(t *testing.T) {
		got := GenerateID(cfg, "db", "refined", "", "")
		want := "dpl_ane2_db_dev_refined_sales_api"
		if got != want {
			t.Errorf("GenerateID() = %q, want %q", got, want)
		}
	})

	t.Run("explicit delimiter overrides config", func(t *testing.T) {
		got := GenerateID(cfg, "sg", "emr", "", "-")
		want := "dpl-ane2-sg-dev-emr-sales-api"
		if got != want {
			t.Errorf("GenerateID() = %q, want %q", got, want)
		}
	})
}

func TestGenerateID_RealWorldWorkspaces(t *testing.T) {
	tests := []struct {
		name         string
		workspace    string
		stage        string
		resourceType string
		qualifier    string
		instanceKey  string
		delimiter    string
		want         string
	}{
		// --- Single-segment workspaces ---
		{
			name:         "vpc/sg",
			workspace:    "vpc",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev-vpc",
		},
		{
			name:         "vpc/sbn with qualifier and instance key",
			workspace:    "vpc",
			resourceType: "sbn",
			qualifier:    "pri",
			instanceKey:  "01",
			want:         "dpl-ane2-sbn-dev-pri-vpc-01",
		},
		{
			name:         "vpc/rtb",
			workspace:    "vpc",
			resourceType: "rtb",
			qualifier:    "pri",
			want:         "dpl-ane2-rtb-dev-pri-vpc",
		},
		{
			name:         "vpc/igw",
			workspace:    "vpc",
			resourceType: "igw",
			want:         "dpl-ane2-igw-dev-vpc",
		},
		{
			name:         "core/kms",
			workspace:    "core",
			resourceType: "kms",
			qualifier:    "s3",
			want:         "dpl-ane2-kms-dev-s3-core",
		},
		{
			name:         "lake/s3 raw-std",
			workspace:    "lake",
			resourceType: "s3",
			qualifier:    "raw-std",
			want:         "dpl-ane2-s3-dev-raw-std-lake",
		},
		{
			name:         "lake/db glue with underscore",
			workspace:    "lake",
			resourceType: "db",
			qualifier:    "refined-std",
			delimiter:    "_",
			want:         "dpl_ane2_db_dev_refined-std_lake",
		},
		{
			name:         "msk/sg",
			workspace:    "msk",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev-msk",
		},
		{
			name:         "ecr/ecr with instance key",
			workspace:    "ecr",
			resourceType: "ecr",
			instanceKey:  "spark-etl",
			want:         "dpl-ane2-ecr-dev-ecr-spark-etl",
		},
		// --- Two-segment workspaces ---
		{
			name:         "sales-api/emr shared-pii",
			workspace:    "sales-api",
			resourceType: "emr",
			instanceKey:  "shared-pii",
			want:         "dpl-ane2-emr-dev-sales-api-shared-pii",
		},
		{
			name:         "sales-api/sg emr",
			workspace:    "sales-api",
			resourceType: "sg",
			qualifier:    "emr",
			want:         "dpl-ane2-sg-dev-emr-sales-api",
		},
		{
			name:         "sales-api/role emr with instance key",
			workspace:    "sales-api",
			resourceType: "role",
			qualifier:    "emr",
			instanceKey:  "shared-pii-etl",
			want:         "dpl-ane2-role-dev-emr-sales-api-shared-pii-etl",
		},
		{
			name:         "hr-web/sg emr",
			workspace:    "hr-web",
			resourceType: "sg",
			qualifier:    "emr",
			want:         "dpl-ane2-sg-dev-emr-hr-web",
		},
		// --- Three-segment workspaces ---
		{
			name:         "sales-api-orders/dmsrt",
			workspace:    "sales-api-orders",
			resourceType: "dmsrt",
			want:         "dpl-ane2-dmsrt-dev-sales-api-orders",
		},
		{
			name:         "sales-api-orders/dmsep src",
			workspace:    "sales-api-orders",
			resourceType: "dmsep",
			qualifier:    "src",
			want:         "dpl-ane2-dmsep-dev-src-sales-api-orders",
		},
		// --- Workspace with underscore ---
		{
			name:         "eks-v1_34/sg",
			workspace:    "eks-v1_34",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev-eks-v1_34",
		},
		// --- Empty workspace ---
		{
			name:         "empty workspace/sg",
			workspace:    "",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev",
		},
		// --- PRD stage ---
		{
			name:         "prd stage sales-api/sg",
			workspace:    "sales-api",
			stage:        "prd",
			resourceType: "sg",
			qualifier:    "emr",
			want:         "dpl-ane2-sg-prd-emr-sales-api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage := tt.stage
			if stage == "" {
				stage = "dev"
			}
			cfg := &LabelConfig{
				Tenant:      "dpl",
				Environment: "ane2",
				Stage:       stage,
				Workspace:   tt.workspace,
				Namespace:   "acme",
				Delimiter:   "-",
			}
			got := GenerateID(cfg, tt.resourceType, tt.qualifier, tt.instanceKey, tt.delimiter)
			if got != tt.want {
				t.Errorf("GenerateID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateTags_RealWorldPatterns(t *testing.T) {
	tests := []struct {
		name           string
		workspace      string
		resourceType   string
		qualifier      string
		instanceKey    string
		delimiter      string
		wantName       string
		wantAttributes string
		wantHasAttrs   bool
	}{
		{
			name:           "sales-api/sg emr",
			workspace:      "sales-api",
			resourceType:   "sg",
			qualifier:      "emr",
			wantName:       "dpl-ane2-sg-dev-emr-sales-api",
			wantAttributes: "emr-sales-api",
			wantHasAttrs:   true,
		},
		{
			name:           "empty workspace/sg - no workspace in attrs",
			workspace:      "",
			resourceType:   "sg",
			wantName:       "dpl-ane2-sg-dev",
			wantHasAttrs:   false,
		},
		{
			name:           "sales-api-orders/dmsep src - 3-part workspace",
			workspace:      "sales-api-orders",
			resourceType:   "dmsep",
			qualifier:      "src",
			wantName:       "dpl-ane2-dmsep-dev-src-sales-api-orders",
			wantAttributes: "src-sales-api-orders",
			wantHasAttrs:   true,
		},
		{
			name:           "sales-api/role emr instance key",
			workspace:      "sales-api",
			resourceType:   "role",
			qualifier:      "emr",
			instanceKey:    "shared-pii-etl",
			wantName:       "dpl-ane2-role-dev-emr-sales-api-shared-pii-etl",
			wantAttributes: "emr-sales-api-shared-pii-etl",
			wantHasAttrs:   true,
		},
		{
			name:           "lake/db underscore delimiter",
			workspace:      "lake",
			resourceType:   "db",
			qualifier:      "refined-std",
			delimiter:      "_",
			wantName:       "dpl_ane2_db_dev_refined-std_lake",
			wantAttributes: "refined-std-lake",
			wantHasAttrs:   true,
		},
		{
			name:           "vpc/sg single-segment workspace",
			workspace:      "vpc",
			resourceType:   "sg",
			wantName:       "dpl-ane2-sg-dev-vpc",
			wantAttributes: "vpc",
			wantHasAttrs:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &LabelConfig{
				Tenant:      "dpl",
				Environment: "ane2",
				Stage:       "dev",
				Workspace:   tt.workspace,
				Namespace:   "acme",
				Delimiter:   "-",
			}
			tags := GenerateTags(cfg, tt.resourceType, tt.qualifier, tt.instanceKey, tt.delimiter)

			if tags["Name"] != tt.wantName {
				t.Errorf("Name = %q, want %q", tags["Name"], tt.wantName)
			}

			attrs, hasAttrs := tags["Attributes"]
			if tt.wantHasAttrs {
				if !hasAttrs {
					t.Error("expected Attributes key in tags")
				} else if attrs != tt.wantAttributes {
					t.Errorf("Attributes = %q, want %q", attrs, tt.wantAttributes)
				}
			} else {
				if hasAttrs {
					t.Errorf("unexpected Attributes key in tags: %q", attrs)
				}
			}
		})
	}
}
