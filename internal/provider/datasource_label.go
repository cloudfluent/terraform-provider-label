package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = (*LabelDataSource)(nil)

type LabelDataSource struct {
	config *LabelConfig
}

type LabelDataSourceModel struct {
	ResourceType    types.String `tfsdk:"resource_type"`
	Qualifier       types.String `tfsdk:"qualifier"`
	InstanceKey     types.String `tfsdk:"instance_key"`
	Delimiter       types.String `tfsdk:"delimiter"`
	Id              types.String `tfsdk:"id"`
	Tags            types.Map    `tfsdk:"tags"`
	TagsWithoutName types.Map    `tfsdk:"tags_without_name"`
}

func NewLabelDataSource() datasource.DataSource {
	return &LabelDataSource{}
}

func (d *LabelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName
}

func (d *LabelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates a consistent resource ID and tags.",
		Attributes: map[string]schema.Attribute{
			"resource_type": schema.StringAttribute{
				Required:    true,
				Description: "The resource type abbreviation (e.g. sg, role, emr, db)",
			},
			"qualifier": schema.StringAttribute{
				Optional:    true,
				Description: "Qualifier segment (e.g. emr, msk)",
			},
			"instance_key": schema.StringAttribute{
				Optional:    true,
				Description: "Instance key for distinguishing multiple resources of the same type",
			},
			"delimiter": schema.StringAttribute{
				Optional:    true,
				Description: "Override the provider-level delimiter for this resource",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Generated resource identifier",
			},
			"tags": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Generated resource tags (includes Name)",
			},
			"tags_without_name": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Generated resource tags without Name key",
			},
		},
	}
}

func (d *LabelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*LabelConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *LabelConfig, got: %T", req.ProviderData),
		)
		return
	}

	d.config = cfg
}

func (d *LabelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model LabelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.config == nil {
		resp.Diagnostics.AddError(
			"Provider Not Configured",
			"The label provider must be configured with tenant, environment, stage, and workspace.",
		)
		return
	}

	var missing []string
	if d.config.Tenant == "" {
		missing = append(missing, "tenant")
	}
	if d.config.Environment == "" {
		missing = append(missing, "environment")
	}
	if d.config.Stage == "" {
		missing = append(missing, "stage")
	}
	if d.config.Workspace == "" {
		missing = append(missing, "workspace")
	}
	if len(missing) > 0 {
		resp.Diagnostics.AddError(
			"Incomplete Provider Configuration",
			fmt.Sprintf("Missing required provider values: %s", strings.Join(missing, ", ")),
		)
		return
	}

	resourceType := model.ResourceType.ValueString()

	var qualifier, instanceKey, delimiter string
	if !model.Qualifier.IsNull() {
		qualifier = model.Qualifier.ValueString()
	}
	if !model.InstanceKey.IsNull() {
		instanceKey = model.InstanceKey.ValueString()
	}
	if !model.Delimiter.IsNull() {
		delimiter = model.Delimiter.ValueString()
	}

	id := GenerateID(d.config, resourceType, qualifier, instanceKey, delimiter)
	tags := GenerateTags(d.config, resourceType, qualifier, instanceKey, delimiter)

	model.Id = types.StringValue(id)

	tagsMap, diags := types.MapValueFrom(ctx, types.StringType, tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Tags = tagsMap

	tagsNoName := make(map[string]string, len(tags))
	for k, v := range tags {
		if k != "Name" {
			tagsNoName[k] = v
		}
	}
	tagsNoNameMap, diags := types.MapValueFrom(ctx, types.StringType, tagsNoName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.TagsWithoutName = tagsNoNameMap

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
