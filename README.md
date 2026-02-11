# terraform-provider-label

Organizations managing complex, multi-account AWS infrastructure need consistent naming and tagging across hundreds of resources. Without a single source of truth, resource names drift, tags become inconsistent, and operational overhead grows.

`terraform-provider-label` solves this by encoding your organization's naming convention as a Terraform provider. Define your naming components once at the provider level, and every resource inherits them automatically — no copy-paste, no drift, no extra state.

Inspired by [cloudposse/terraform-null-label](https://github.com/cloudposse/terraform-null-label), this provider addresses practical limitations encountered when using module-based approaches at scale:

- Every module instance creates state entries, even though naming is a pure computation
- Provider-level values (tenant, environment, stage) must be passed explicitly to every module call
- CI/CD tools that abstract workspaces have no clean way to inject naming context

## Features

- **Zero state** — all values are computed at plan time, nothing is stored in Terraform state
- **Provider-level defaults** — define tenant, environment, stage, and workspace once; every data source inherits them
- **Per-resource overrides** — customize delimiter, qualifier, or instance key for individual resources
- **Consistent tags** — automatically generates `Name`, `Tenant`, `Environment`, `Stage`, `Namespace`, and `Attributes`
- **`for_each` friendly** — create multiple labels of the same resource type in a single block

## Installation

```hcl
terraform {
  required_providers {
    label = {
      source  = "cloudfluent/label"
      version = "~> 0.1"
    }
  }
}
```

## Usage

### Provider Configuration

```hcl
provider "label" {
  tenant      = "dpl"
  environment = "ane2"
  stage       = "dev"
  workspace   = "data-sales-api"
  namespace   = "acme"
  # delimiter = "-"  # default
}
```

All provider attributes are optional and fall back to environment variables:

| Attribute | Environment Variable |
|-----------|---------------------|
| `tenant` | `LABEL_TENANT` |
| `environment` | `LABEL_ENVIRONMENT` |
| `stage` | `LABEL_STAGE` |
| `workspace` | `LABEL_WORKSPACE` |
| `namespace` | `LABEL_NAMESPACE` |
| `delimiter` | `LABEL_DELIMITER` |

### CI/CD Integration

Organizations using external Terraform CI/CD tools (Scalr, Terraform Cloud, Spacelift, etc.) that manage workspaces as an abstraction layer can inject all naming components via environment variables. This keeps the Terraform code workspace-agnostic — the same configuration deploys to any environment without modification.

```hcl
# No hardcoded values — everything comes from the CI/CD environment
provider "label" {}
```

The CI/CD tool sets the environment variables per workspace:

```bash
# dev workspace
LABEL_TENANT=dpl  LABEL_ENVIRONMENT=ane2  LABEL_STAGE=dev  LABEL_WORKSPACE=data-sales-api  LABEL_NAMESPACE=acme

# prd workspace
LABEL_TENANT=dpl  LABEL_ENVIRONMENT=ane2  LABEL_STAGE=prd  LABEL_WORKSPACE=data-sales-api  LABEL_NAMESPACE=acme
```

This way, `dev` and `prd` share identical `.tf` files and only differ by the injected variables.

### Data Source

```hcl
data "label" "sg" {
  resource_type = "sg"
  qualifier     = "emr"
}

resource "aws_security_group" "emr" {
  name = data.label.sg.id              # dpl-ane2-sg-dev-emr-sales-api
  tags = data.label.sg.tags_without_name
}
```

### ID Format

```
{tenant}-{environment}-{resource_type}-{stage}-{qualifier}-{domain}-{instance_key}
```

The **domain** is extracted from the `workspace` value by dropping the first segment:

| workspace | domain |
|-----------|--------|
| `data-sales-api` | `sales-api` |
| `dms-sales-api-orders` | `sales-api-orders` |
| `eks-v1_34` | `v1_34` |
| `vpc` | _(none)_ |

### Examples

```hcl
# Simple
data "label" "sg" {
  resource_type = "sg"
}
# => dpl-ane2-sg-dev-sales-api

# With qualifier
data "label" "sg_emr" {
  resource_type = "sg"
  qualifier     = "emr"
}
# => dpl-ane2-sg-dev-emr-sales-api

# Qualifier + instance_key
data "label" "role_emr" {
  resource_type = "role"
  qualifier     = "emr"
  instance_key  = "shared-pii-etl"
}
# => dpl-ane2-role-dev-emr-sales-api-shared-pii-etl

# Delimiter override (e.g. Glue database)
data "label" "db" {
  resource_type = "db"
  qualifier     = "refined"
  delimiter     = "_"
}
# => dpl_ane2_db_dev_refined_sales_api

# Multiple resources via for_each
data "label" "sgs" {
  for_each      = toset(["emr", "msk", "vpce"])
  resource_type = "sg"
  qualifier     = each.value
}
```

### Tags Output

`tags` includes a `Name` key. `tags_without_name` excludes it, useful when the resource sets `name` as a separate argument.

```hcl
data "label" "sg_emr" {
  resource_type = "sg"
  qualifier     = "emr"
}

# data.label.sg_emr.tags =>
# {
#   Name        = "dpl-ane2-sg-dev-emr-sales-api"
#   Tenant      = "dpl"
#   Environment = "ane2"
#   Stage       = "dev"
#   Namespace   = "acme"
#   Attributes  = "emr-sales-api"
# }
```

## Development

```bash
make build       # Build the provider binary
make test        # Run all tests
make testacc     # Run acceptance tests
make lint        # Format and vet
make docs        # Regenerate registry documentation
```

## License

[MPL-2.0](LICENSE)
