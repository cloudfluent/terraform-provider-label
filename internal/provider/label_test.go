package provider

import (
	"testing"
)

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		workspace string
		want      []string
	}{
		{"data-sales-api", []string{"sales", "api"}},
		{"data-hr-web", []string{"hr", "web"}},
		{"data-ops-svc", []string{"ops", "svc"}},
		{"dms-sales-api-orders", []string{"sales", "api", "orders"}},
		{"vpc", nil},
		{"eks-v1_34", []string{"v1_34"}},
		{"rds-postgres", []string{"postgres"}},
	}

	for _, tt := range tests {
		t.Run(tt.workspace, func(t *testing.T) {
			got := ExtractDomain(tt.workspace)
			if len(got) != len(tt.want) {
				t.Fatalf("ExtractDomain(%q) = %v, want %v", tt.workspace, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ExtractDomain(%q)[%d] = %q, want %q", tt.workspace, i, got[i], tt.want[i])
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
		Workspace:   "data-sales-api",
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
			name:         "vpc workspace (no domain)",
			workspace:    "vpc",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev",
		},
		{
			name:         "dms workspace with extra parts",
			workspace:    "dms-sales-api-orders",
			resourceType: "task",
			want:         "dpl-ane2-task-dev-sales-api-orders",
		},
		{
			name:         "eks workspace",
			workspace:    "eks-v1_34",
			resourceType: "node",
			want:         "dpl-ane2-node-dev-v1_34",
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
		Workspace:   "data-sales-api",
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
		Workspace:   "data-sales-api",
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
		Workspace:   "data-sales-api",
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
		Workspace:   "data-sales-api",
		Delimiter:   "_",
	}

	t.Run("env delimiter used as default", func(t *testing.T) {
		got := GenerateID(cfg, "db", "refined", "", "")
		want := "dpl_ane2_db_dev_refined_sales_api"
		if got != want {
			t.Errorf("GenerateID() = %q, want %q", got, want)
		}
	})

	t.Run("explicit delimiter overrides env", func(t *testing.T) {
		got := GenerateID(cfg, "sg", "emr", "", "-")
		want := "dpl-ane2-sg-dev-emr-sales-api"
		if got != want {
			t.Errorf("GenerateID() = %q, want %q", got, want)
		}
	})
}

// TestGenerateID_RealWorldWorkspaces tests naming across all real workspace types.
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
		// --- VPC workspace (no domain) ---
		{
			name:         "vpc/sg simple",
			workspace:    "vpc",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev",
		},
		{
			name:         "vpc/sbn with qualifier",
			workspace:    "vpc",
			resourceType: "sbn",
			qualifier:    "pri",
			instanceKey:  "01",
			want:         "dpl-ane2-sbn-dev-pri-01",
		},
		{
			name:         "vpc/rtb",
			workspace:    "vpc",
			resourceType: "rtb",
			qualifier:    "pri",
			want:         "dpl-ane2-rtb-dev-pri",
		},
		{
			name:         "vpc/igw",
			workspace:    "vpc",
			resourceType: "igw",
			want:         "dpl-ane2-igw-dev",
		},
		{
			name:         "vpc/nacl",
			workspace:    "vpc",
			resourceType: "nacl",
			qualifier:    "pri",
			want:         "dpl-ane2-nacl-dev-pri",
		},
		{
			name:         "vpc/vpce",
			workspace:    "vpc",
			resourceType: "vpce",
			qualifier:    "s3",
			want:         "dpl-ane2-vpce-dev-s3",
		},
		{
			name:         "vpc/tgwattach",
			workspace:    "vpc",
			resourceType: "tgwattach",
			want:         "dpl-ane2-tgwattach-dev",
		},
		{
			name:         "vpc/pl prefix list",
			workspace:    "vpc",
			resourceType: "pl",
			qualifier:    "gov-tgw",
			want:         "dpl-ane2-pl-dev-gov-tgw",
		},
		// --- Core workspace ---
		{
			name:         "core/kms",
			workspace:    "core",
			resourceType: "kms",
			qualifier:    "s3",
			want:         "dpl-ane2-kms-dev-s3",
		},
		{
			name:         "core/s3",
			workspace:    "core",
			resourceType: "s3",
			qualifier:    "logs",
			want:         "dpl-ane2-s3-dev-logs",
		},
		// --- Lake workspace ---
		{
			name:         "lake/s3 raw-std",
			workspace:    "lake",
			resourceType: "s3",
			qualifier:    "raw-std",
			want:         "dpl-ane2-s3-dev-raw-std",
		},
		{
			name:         "lake/s3 refined-pii",
			workspace:    "lake",
			resourceType: "s3",
			qualifier:    "refined-pii",
			want:         "dpl-ane2-s3-dev-refined-pii",
		},
		{
			name:         "lake/db glue with underscore",
			workspace:    "lake",
			resourceType: "db",
			qualifier:    "refined-std",
			delimiter:    "_",
			want:         "dpl_ane2_db_dev_refined-std",
		},
		// --- data workspace (2-part domain) ---
		{
			name:         "data-sales-api/emr shared-pii",
			workspace:    "data-sales-api",
			resourceType: "emr",
			instanceKey:  "shared-pii",
			want:         "dpl-ane2-emr-dev-sales-api-shared-pii",
		},
		{
			name:         "data-sales-api/emr shared-std",
			workspace:    "data-sales-api",
			resourceType: "emr",
			instanceKey:  "shared-std",
			want:         "dpl-ane2-emr-dev-sales-api-shared-std",
		},
		{
			name:         "data-sales-api/role emr with long instance key",
			workspace:    "data-sales-api",
			resourceType: "role",
			qualifier:    "emr",
			instanceKey:  "shared-pii-etl",
			want:         "dpl-ane2-role-dev-emr-sales-api-shared-pii-etl",
		},
		{
			name:         "data-sales-api/sg emr",
			workspace:    "data-sales-api",
			resourceType: "sg",
			qualifier:    "emr",
			want:         "dpl-ane2-sg-dev-emr-sales-api",
		},
		{
			name:         "data-sales-api/sg msk",
			workspace:    "data-sales-api",
			resourceType: "sg",
			qualifier:    "msk",
			want:         "dpl-ane2-sg-dev-msk-sales-api",
		},
		// --- data-hr-web workspace ---
		{
			name:         "data-hr-web/emr shared-pii",
			workspace:    "data-hr-web",
			resourceType: "emr",
			instanceKey:  "shared-pii",
			want:         "dpl-ane2-emr-dev-hr-web-shared-pii",
		},
		{
			name:         "data-hr-web/sg emr",
			workspace:    "data-hr-web",
			resourceType: "sg",
			qualifier:    "emr",
			want:         "dpl-ane2-sg-dev-emr-hr-web",
		},
		// --- data-ops-svc workspace ---
		{
			name:         "data-ops-svc/sg vpce",
			workspace:    "data-ops-svc",
			resourceType: "sg",
			qualifier:    "vpce",
			want:         "dpl-ane2-sg-dev-vpce-ops-svc",
		},
		// --- DMS workspaces (multi-part domain) ---
		{
			name:         "dms-sales-api-orders/dmsrt",
			workspace:    "dms-sales-api-orders",
			resourceType: "dmsrt",
			want:         "dpl-ane2-dmsrt-dev-sales-api-orders",
		},
		{
			name:         "dms-sales-api-orders/dmsep",
			workspace:    "dms-sales-api-orders",
			resourceType: "dmsep",
			qualifier:    "src",
			want:         "dpl-ane2-dmsep-dev-src-sales-api-orders",
		},
		{
			name:         "dms-sales-api-users/dmsrt",
			workspace:    "dms-sales-api-users",
			resourceType: "dmsrt",
			want:         "dpl-ane2-dmsrt-dev-sales-api-users",
		},
		{
			name:         "dms-sales-api-products/dmsrt",
			workspace:    "dms-sales-api-products",
			resourceType: "dmsrt",
			want:         "dpl-ane2-dmsrt-dev-sales-api-products",
		},
		// --- EKS workspace (underscore in domain) ---
		{
			name:         "eks-v1_34/sg",
			workspace:    "eks-v1_34",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev-v1_34",
		},
		{
			name:         "eks-v1_34/ec2 nodegroup",
			workspace:    "eks-v1_34",
			resourceType: "ec2",
			qualifier:    "nodegroup",
			want:         "dpl-ane2-ec2-dev-nodegroup-v1_34",
		},
		// --- RDS workspace ---
		{
			name:         "rds-postgres/dbpg",
			workspace:    "rds-postgres",
			resourceType: "dbpg",
			qualifier:    "trino-gw",
			want:         "dpl-ane2-dbpg-dev-trino-gw-postgres",
		},
		{
			name:         "rds-postgres/dbcpg",
			workspace:    "rds-postgres",
			resourceType: "dbcpg",
			qualifier:    "superset",
			want:         "dpl-ane2-dbcpg-dev-superset-postgres",
		},
		// --- MSK workspace ---
		{
			name:         "msk/sg",
			workspace:    "msk",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev",
		},
		// --- MWAA workspace ---
		{
			name:         "mwaa/sg",
			workspace:    "mwaa",
			resourceType: "sg",
			want:         "dpl-ane2-sg-dev",
		},
		// --- ECR workspace ---
		{
			name:         "ecr/ecr with instance key",
			workspace:    "ecr",
			resourceType: "ecr",
			instanceKey:  "spark-etl",
			want:         "dpl-ane2-ecr-dev-spark-etl",
		},
		// --- Access workspace ---
		{
			name:         "access/role",
			workspace:    "access",
			resourceType: "role",
			qualifier:    "emr",
			instanceKey:  "shared-pii-etl",
			want:         "dpl-ane2-role-dev-emr-shared-pii-etl",
		},
		{
			name:         "access/policy",
			workspace:    "access",
			resourceType: "policy",
			qualifier:    "s3",
			instanceKey:  "raw-std-read",
			want:         "dpl-ane2-policy-dev-s3-raw-std-read",
		},
		// --- PRD stage ---
		{
			name:         "prd stage data-sales-api/sg",
			workspace:    "data-sales-api",
			stage:        "prd",
			resourceType: "sg",
			qualifier:    "emr",
			want:         "dpl-ane2-sg-prd-emr-sales-api",
		},
		{
			name:         "prd stage vpc/sbn",
			workspace:    "vpc",
			stage:        "prd",
			resourceType: "sbn",
			qualifier:    "pri",
			instanceKey:  "01",
			want:         "dpl-ane2-sbn-prd-pri-01",
		},
		// --- Route53 workspace ---
		{
			name:         "route53/simple",
			workspace:    "route53",
			resourceType: "zone",
			want:         "dpl-ane2-zone-dev",
		},
		// --- Secrets Manager workspace ---
		{
			name:         "secretsmanager/secrets",
			workspace:    "secretsmanager",
			resourceType: "secrets",
			qualifier:    "rds",
			instanceKey:  "trino-gw",
			want:         "dpl-ane2-secrets-dev-rds-trino-gw",
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

// TestGenerateTags_RealWorldPatterns tests tag generation for diverse scenarios.
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
		wantNamespace  bool
	}{
		{
			name:           "data-sales-api/sg emr - full tags",
			workspace:      "data-sales-api",
			resourceType:   "sg",
			qualifier:      "emr",
			wantName:       "dpl-ane2-sg-dev-emr-sales-api",
			wantAttributes: "emr-sales-api",
			wantHasAttrs:   true,
			wantNamespace:  true,
		},
		{
			name:         "vpc/sg no qualifier - domain-less",
			workspace:    "vpc",
			resourceType: "sg",
			wantName:     "dpl-ane2-sg-dev",
			wantHasAttrs: false,
			wantNamespace: true,
		},
		{
			name:           "dms-sales-api-orders/dmsep - 3-part domain in attributes",
			workspace:      "dms-sales-api-orders",
			resourceType:   "dmsep",
			qualifier:      "src",
			wantName:       "dpl-ane2-dmsep-dev-src-sales-api-orders",
			wantAttributes: "src-sales-api-orders",
			wantHasAttrs:   true,
			wantNamespace:  true,
		},
		{
			name:           "data-sales-api/role emr instance key - complex attributes",
			workspace:      "data-sales-api",
			resourceType:   "role",
			qualifier:      "emr",
			instanceKey:    "shared-pii-etl",
			wantName:       "dpl-ane2-role-dev-emr-sales-api-shared-pii-etl",
			wantAttributes: "emr-sales-api-shared-pii-etl",
			wantHasAttrs:   true,
			wantNamespace:  true,
		},
		{
			name:           "lake/db underscore delimiter - name uses underscore, attrs use dash",
			workspace:      "lake",
			resourceType:   "db",
			qualifier:      "refined-std",
			delimiter:      "_",
			wantName:       "dpl_ane2_db_dev_refined-std",
			wantAttributes: "refined-std",
			wantHasAttrs:   true,
			wantNamespace:  true,
		},
		{
			name:           "ecr/ecr instance key only - no qualifier",
			workspace:      "ecr",
			resourceType:   "ecr",
			instanceKey:    "spark-etl",
			wantName:       "dpl-ane2-ecr-dev-spark-etl",
			wantAttributes: "spark-etl",
			wantHasAttrs:   true,
			wantNamespace:  true,
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
			if tags["Tenant"] != "dpl" {
				t.Errorf("Tenant = %q, want %q", tags["Tenant"], "dpl")
			}
			if tags["Environment"] != "ane2" {
				t.Errorf("Environment = %q, want %q", tags["Environment"], "ane2")
			}
			if tags["Stage"] != "dev" {
				t.Errorf("Stage = %q, want %q", tags["Stage"], "dev")
			}

			if tt.wantNamespace {
				if tags["Namespace"] != "acme" {
					t.Errorf("Namespace = %q, want %q", tags["Namespace"], "acme")
				}
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

// TestExtractDomain_AllWorkspaces covers all workspace types.
func TestExtractDomain_AllWorkspaces(t *testing.T) {
	tests := []struct {
		workspace string
		want      []string
	}{
		// No domain (single-segment)
		{"vpc", nil},
		{"core", nil},
		{"lake", nil},
		{"msk", nil},
		{"mwaa", nil},
		{"ecr", nil},
		{"access", nil},
		{"route53", nil},
		{"secretsmanager", nil},
		// 2-part domain
		{"data-sales-api", []string{"sales", "api"}},
		{"data-hr-web", []string{"hr", "web"}},
		{"data-ops-svc", []string{"ops", "svc"}},
		// 1-part domain
		{"eks-v1_34", []string{"v1_34"}},
		{"rds-postgres", []string{"postgres"}},
		// 3+ part domain (DMS)
		{"dms-sales-api-orders", []string{"sales", "api", "orders"}},
		{"dms-sales-api-users", []string{"sales", "api", "users"}},
		{"dms-sales-api-products", []string{"sales", "api", "products"}},
	}

	for _, tt := range tests {
		t.Run(tt.workspace, func(t *testing.T) {
			got := ExtractDomain(tt.workspace)
			if len(got) != len(tt.want) {
				t.Fatalf("ExtractDomain(%q) = %v (len=%d), want %v (len=%d)",
					tt.workspace, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ExtractDomain(%q)[%d] = %q, want %q",
						tt.workspace, i, got[i], tt.want[i])
				}
			}
		})
	}
}
