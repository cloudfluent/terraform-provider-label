# Basic usage
data "label" "sg" {
  resource_type = "sg"
}
# => id: dpl-ane2-sg-dev-sales-api

# With qualifier
data "label" "sg_emr" {
  resource_type = "sg"
  qualifier     = "emr"
}
# => id: dpl-ane2-sg-dev-emr-sales-api

# Qualifier + instance_key
data "label" "role_emr" {
  resource_type = "role"
  qualifier     = "emr"
  instance_key  = "shared-pii-etl"
}
# => id: dpl-ane2-role-dev-emr-sales-api-shared-pii-etl

# Delimiter override
data "label" "db" {
  resource_type = "db"
  qualifier     = "refined"
  delimiter     = "_"
}
# => id: dpl_ane2_db_dev_refined_sales_api

# Multiple resources of the same type via for_each
data "label" "sgs" {
  for_each      = toset(["emr", "msk", "vpce"])
  resource_type = "sg"
  qualifier     = each.value
}

# Usage example
resource "aws_security_group" "emr" {
  name = data.label.sg_emr.id
  tags = data.label.sg_emr.tags_without_name
}
