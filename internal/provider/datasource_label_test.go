package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

const testProviderConfig = `
terraform {
  required_providers {
    label = {
      source = "cloudfluent/label"
    }
  }
}
`

var testProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"label": providerserver.NewProtocol6WithError(New()),
}

const testProviderConfigWithValues = testProviderConfig + `
provider "label" {
  tenant      = "dpl"
  environment = "ane2"
  stage       = "dev"
  workspace   = "sales-api"
  namespace   = "acme"
}
`

func TestLabelDataSource_Simple(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfigWithValues + `
data "label" "test" {
  resource_type = "sg"
}

output "id" {
  value = data.label.test.id
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("id", knownvalue.StringExact("dpl-ane2-sg-dev-sales-api")),
				},
			},
		},
	})
}

func TestLabelDataSource_WithQualifier(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfigWithValues + `
data "label" "test" {
  resource_type = "sg"
  qualifier     = "emr"
}

output "id" {
  value = data.label.test.id
}

output "tags" {
  value = data.label.test.tags
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("id", knownvalue.StringExact("dpl-ane2-sg-dev-emr-sales-api")),
					statecheck.ExpectKnownOutputValue("tags", knownvalue.MapExact(map[string]knownvalue.Check{
						"Name":        knownvalue.StringExact("dpl-ane2-sg-dev-emr-sales-api"),
						"Namespace":   knownvalue.StringExact("acme"),
						"Tenant":      knownvalue.StringExact("dpl"),
						"Environment": knownvalue.StringExact("ane2"),
						"Stage":       knownvalue.StringExact("dev"),
						"Attributes":  knownvalue.StringExact("emr-sales-api"),
					})),
				},
			},
		},
	})
}

func TestLabelDataSource_WithInstanceKey(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfigWithValues + `
data "label" "test" {
  resource_type = "role"
  qualifier     = "emr"
  instance_key  = "shared-pii-etl"
}

output "id" {
  value = data.label.test.id
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("id", knownvalue.StringExact("dpl-ane2-role-dev-emr-sales-api-shared-pii-etl")),
				},
			},
		},
	})
}

func TestLabelDataSource_CustomDelimiter(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfigWithValues + `
data "label" "test" {
  resource_type = "db"
  qualifier     = "refined"
  delimiter     = "_"
}

output "id" {
  value = data.label.test.id
}

output "tags" {
  value = data.label.test.tags
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("id", knownvalue.StringExact("dpl_ane2_db_dev_refined_sales_api")),
					statecheck.ExpectKnownOutputValue("tags", knownvalue.MapExact(map[string]knownvalue.Check{
						"Name":        knownvalue.StringExact("dpl_ane2_db_dev_refined_sales_api"),
						"Namespace":   knownvalue.StringExact("acme"),
						"Tenant":      knownvalue.StringExact("dpl"),
						"Environment": knownvalue.StringExact("ane2"),
						"Stage":       knownvalue.StringExact("dev"),
						"Attributes":  knownvalue.StringExact("refined-sales-api"),
					})),
				},
			},
		},
	})
}

func TestLabelDataSource_TagsWithoutName(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfigWithValues + `
data "label" "test" {
  resource_type = "sg"
  qualifier     = "emr"
}

output "tags" {
  value = data.label.test.tags
}

output "tags_without_name" {
  value = data.label.test.tags_without_name
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("tags", knownvalue.MapExact(map[string]knownvalue.Check{
						"Name":        knownvalue.StringExact("dpl-ane2-sg-dev-emr-sales-api"),
						"Namespace":   knownvalue.StringExact("acme"),
						"Tenant":      knownvalue.StringExact("dpl"),
						"Environment": knownvalue.StringExact("ane2"),
						"Stage":       knownvalue.StringExact("dev"),
						"Attributes":  knownvalue.StringExact("emr-sales-api"),
					})),
					statecheck.ExpectKnownOutputValue("tags_without_name", knownvalue.MapExact(map[string]knownvalue.Check{
						"Namespace":   knownvalue.StringExact("acme"),
						"Tenant":      knownvalue.StringExact("dpl"),
						"Environment": knownvalue.StringExact("ane2"),
						"Stage":       knownvalue.StringExact("dev"),
						"Attributes":  knownvalue.StringExact("emr-sales-api"),
					})),
				},
			},
		},
	})
}

func TestLabelDataSource_ProviderDelimiter(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig + `
provider "label" {
  tenant      = "dpl"
  environment = "ane2"
  stage       = "dev"
  workspace   = "sales-api"
  namespace   = "acme"
  delimiter   = "_"
}

data "label" "test" {
  resource_type = "db"
  qualifier     = "refined"
}

output "id" {
  value = data.label.test.id
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("id", knownvalue.StringExact("dpl_ane2_db_dev_refined_sales_api")),
				},
			},
		},
	})
}

// TestLabelDataSource_VpcWorkspace tests single-segment workspace.
func TestLabelDataSource_VpcWorkspace(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig + `
provider "label" {
  tenant      = "dpl"
  environment = "ane2"
  stage       = "dev"
  workspace   = "vpc"
  namespace   = "acme"
}

data "label" "sg" {
  resource_type = "sg"
}

data "label" "sbn" {
  resource_type = "sbn"
  qualifier     = "pri"
  instance_key  = "01"
}

output "sg_id" {
  value = data.label.sg.id
}

output "sbn_id" {
  value = data.label.sbn.id
}

output "sg_tags" {
  value = data.label.sg.tags
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("sg_id", knownvalue.StringExact("dpl-ane2-sg-dev-vpc")),
					statecheck.ExpectKnownOutputValue("sbn_id", knownvalue.StringExact("dpl-ane2-sbn-dev-pri-vpc-01")),
					statecheck.ExpectKnownOutputValue("sg_tags", knownvalue.MapExact(map[string]knownvalue.Check{
						"Name":        knownvalue.StringExact("dpl-ane2-sg-dev-vpc"),
						"Namespace":   knownvalue.StringExact("acme"),
						"Tenant":      knownvalue.StringExact("dpl"),
						"Environment": knownvalue.StringExact("ane2"),
						"Stage":       knownvalue.StringExact("dev"),
						"Attributes":  knownvalue.StringExact("vpc"),
					})),
				},
			},
		},
	})
}

// TestLabelDataSource_MultiSegmentWorkspace tests workspace with 3+ segments.
func TestLabelDataSource_MultiSegmentWorkspace(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig + `
provider "label" {
  tenant      = "dpl"
  environment = "ane2"
  stage       = "dev"
  workspace   = "sales-api-orders"
  namespace   = "acme"
}

data "label" "dmsrt" {
  resource_type = "dmsrt"
}

data "label" "dmsep_src" {
  resource_type = "dmsep"
  qualifier     = "src"
}

output "dmsrt_id" {
  value = data.label.dmsrt.id
}

output "dmsep_src_id" {
  value = data.label.dmsep_src.id
}

output "dmsep_tags" {
  value = data.label.dmsep_src.tags
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("dmsrt_id", knownvalue.StringExact("dpl-ane2-dmsrt-dev-sales-api-orders")),
					statecheck.ExpectKnownOutputValue("dmsep_src_id", knownvalue.StringExact("dpl-ane2-dmsep-dev-src-sales-api-orders")),
					statecheck.ExpectKnownOutputValue("dmsep_tags", knownvalue.MapExact(map[string]knownvalue.Check{
						"Name":        knownvalue.StringExact("dpl-ane2-dmsep-dev-src-sales-api-orders"),
						"Namespace":   knownvalue.StringExact("acme"),
						"Tenant":      knownvalue.StringExact("dpl"),
						"Environment": knownvalue.StringExact("ane2"),
						"Stage":       knownvalue.StringExact("dev"),
						"Attributes":  knownvalue.StringExact("src-sales-api-orders"),
					})),
				},
			},
		},
	})
}

// TestLabelDataSource_PrdStage tests production stage naming.
func TestLabelDataSource_PrdStage(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig + `
provider "label" {
  tenant      = "dpl"
  environment = "ane2"
  stage       = "prd"
  workspace   = "sales-api"
  namespace   = "acme"
}

data "label" "sg" {
  resource_type = "sg"
  qualifier     = "emr"
}

output "id" {
  value = data.label.sg.id
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("id", knownvalue.StringExact("dpl-ane2-sg-prd-emr-sales-api")),
				},
			},
		},
	})
}

// TestLabelDataSource_EksWorkspace tests underscore-containing workspace.
func TestLabelDataSource_EksWorkspace(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig + `
provider "label" {
  tenant      = "dpl"
  environment = "ane2"
  stage       = "dev"
  workspace   = "eks-v1_34"
  namespace   = "acme"
}

data "label" "sg" {
  resource_type = "sg"
}

output "id" {
  value = data.label.sg.id
}

output "tags" {
  value = data.label.sg.tags
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("id", knownvalue.StringExact("dpl-ane2-sg-dev-eks-v1_34")),
					statecheck.ExpectKnownOutputValue("tags", knownvalue.MapExact(map[string]knownvalue.Check{
						"Name":        knownvalue.StringExact("dpl-ane2-sg-dev-eks-v1_34"),
						"Namespace":   knownvalue.StringExact("acme"),
						"Tenant":      knownvalue.StringExact("dpl"),
						"Environment": knownvalue.StringExact("ane2"),
						"Stage":       knownvalue.StringExact("dev"),
						"Attributes":  knownvalue.StringExact("eks-v1_34"),
					})),
				},
			},
		},
	})
}

// TestLabelDataSource_InstanceKeyOnly tests instance_key without qualifier.
func TestLabelDataSource_InstanceKeyOnly(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfigWithValues + `
data "label" "test" {
  resource_type = "emr"
  instance_key  = "shared-pii"
}

output "id" {
  value = data.label.test.id
}

output "tags" {
  value = data.label.test.tags
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("id", knownvalue.StringExact("dpl-ane2-emr-dev-sales-api-shared-pii")),
					statecheck.ExpectKnownOutputValue("tags", knownvalue.MapExact(map[string]knownvalue.Check{
						"Name":        knownvalue.StringExact("dpl-ane2-emr-dev-sales-api-shared-pii"),
						"Namespace":   knownvalue.StringExact("acme"),
						"Tenant":      knownvalue.StringExact("dpl"),
						"Environment": knownvalue.StringExact("ane2"),
						"Stage":       knownvalue.StringExact("dev"),
						"Attributes":  knownvalue.StringExact("sales-api-shared-pii"),
					})),
				},
			},
		},
	})
}

// TestLabelDataSource_EmptyWorkspace tests that workspace is optional.
func TestLabelDataSource_EmptyWorkspace(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig + `
provider "label" {
  tenant      = "dpl"
  environment = "ane2"
  stage       = "dev"
}

data "label" "vpc" {
  resource_type = "vpc"
}

data "label" "sbn" {
  resource_type = "sbn"
  instance_key  = "01"
}

output "vpc_id" {
  value = data.label.vpc.id
}

output "sbn_id" {
  value = data.label.sbn.id
}

output "vpc_tags" {
  value = data.label.vpc.tags
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("vpc_id", knownvalue.StringExact("dpl-ane2-vpc-dev")),
					statecheck.ExpectKnownOutputValue("sbn_id", knownvalue.StringExact("dpl-ane2-sbn-dev-01")),
					statecheck.ExpectKnownOutputValue("vpc_tags", knownvalue.MapExact(map[string]knownvalue.Check{
						"Name":        knownvalue.StringExact("dpl-ane2-vpc-dev"),
						"Tenant":      knownvalue.StringExact("dpl"),
						"Environment": knownvalue.StringExact("ane2"),
						"Stage":       knownvalue.StringExact("dev"),
					})),
				},
			},
		},
	})
}
