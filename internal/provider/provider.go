package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*LabelProvider)(nil)

type LabelProvider struct{}

type LabelProviderModel struct {
	Tenant      types.String `tfsdk:"tenant"`
	Environment types.String `tfsdk:"environment"`
	Stage       types.String `tfsdk:"stage"`
	Workspace   types.String `tfsdk:"workspace"`
	Namespace   types.String `tfsdk:"namespace"`
	Delimiter   types.String `tfsdk:"delimiter"`
}

func New() provider.Provider {
	return &LabelProvider{}
}

func (p *LabelProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "label"
}

func (p *LabelProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Encodes your organization's naming convention as a Terraform provider for consistent resource identifiers and tags.",
		Attributes: map[string]schema.Attribute{
			"tenant": schema.StringAttribute{
				Optional:    true,
				Description: "Tenant identifier (e.g. dpl). Falls back to LABEL_TENANT env var.",
			},
			"environment": schema.StringAttribute{
				Optional:    true,
				Description: "Environment identifier (e.g. ane2). Falls back to LABEL_ENVIRONMENT env var.",
			},
			"stage": schema.StringAttribute{
				Optional:    true,
				Description: "Stage (e.g. dev, prd). Falls back to LABEL_STAGE env var.",
			},
			"workspace": schema.StringAttribute{
				Optional:    true,
				Description: "Workspace name for domain extraction (e.g. data-sales-api). Falls back to LABEL_WORKSPACE env var.",
			},
			"namespace": schema.StringAttribute{
				Optional:    true,
				Description: "Namespace for tags (e.g. acme). Falls back to LABEL_NAMESPACE env var.",
			},
			"delimiter": schema.StringAttribute{
				Optional:    true,
				Description: "Default delimiter (default: -). Falls back to LABEL_DELIMITER env var.",
			},
		},
	}
}

func (p *LabelProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var model LabelProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	delimiter := stringValueOrEnv(model.Delimiter, "LABEL_DELIMITER")
	if delimiter == "" {
		delimiter = "-"
	}

	cfg := &LabelConfig{
		Tenant:      stringValueOrEnv(model.Tenant, "LABEL_TENANT"),
		Environment: stringValueOrEnv(model.Environment, "LABEL_ENVIRONMENT"),
		Stage:       stringValueOrEnv(model.Stage, "LABEL_STAGE"),
		Workspace:   stringValueOrEnv(model.Workspace, "LABEL_WORKSPACE"),
		Namespace:   stringValueOrEnv(model.Namespace, "LABEL_NAMESPACE"),
		Delimiter:   delimiter,
	}

	resp.DataSourceData = cfg
}

func (p *LabelProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

func (p *LabelProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewLabelDataSource,
	}
}

func stringValueOrEnv(v types.String, envKey string) string {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueString()
	}
	return os.Getenv(envKey)
}
