package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAwsEc2ImageBuilderInfrastructureConfiguration(_ context.Context) *plugin.Table {
	return &plugin.Table{
		GetMatrixItemFunc: AllRegionsMatrix,
		Name:              "aws_ec2_image_builder_infrastructure_configuration",
		Description:       "AWS EC2 Image Builder Infrastructure Configuration",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("arn"),
			Hydrate:    getEc2ImageBuilderInfrastructureConfiguration,
			Tags:       map[string]string{"service": "imagebuilder", "action": "GetInfrastructureConfiguration"},
		},
		List: &plugin.ListConfig{
			Hydrate: listEc2ImageBuilderInfrastructureConfigurations,
			Tags:    map[string]string{"service": "imagebuilder", "action": "ListInfrastructureConfigurations"},
		},
		HydrateConfig: []plugin.HydrateConfig{
			{
				Func: getEc2ImageBuilderInfrastructureConfiguration,
				Tags: map[string]string{"service": "imagebuilder", "action": "GetInfrastructureConfiguration"},
			},
		},
		Columns: awsRegionalColumns([]*plugin.Column{
			{
				Name:        "arn",
				Description: "The Amazon Resource Name (ARN) of the infrastructure configuration.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The name of the infrastructure configuration.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the infrastructure configuration.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "date_created",
				Description: "The date on which the infrastructure configuration was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "date_updated",
				Description: "The date on which the infrastructure configuration was last updated.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "instance_profile_name",
				Description: "The instance profile of the infrastructure configuration.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "key_pair",
				Description: "The Amazon EC2 key pair of the infrastructure configuration.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getEc2ImageBuilderInfrastructureConfiguration,
			},
			{
				Name:        "sns_topic_arn",
				Description: "The ARN of the SNS topic to which Image Builder sends image build event notifications.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getEc2ImageBuilderInfrastructureConfiguration,
			},
			{
				Name:        "subnet_id",
				Description: "The subnet ID of the infrastructure configuration.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getEc2ImageBuilderInfrastructureConfiguration,
			},
			{
				Name:        "terminate_instance_on_failure",
				Description: "The terminate instance on failure configuration of the infrastructure configuration.",
				Type:        proto.ColumnType_BOOL,
				Hydrate:     getEc2ImageBuilderInfrastructureConfiguration,
			},
			{
				Name:        "instance_metadata_options",
				Description: "The instance metadata option settings for the infrastructure configuration.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getEc2ImageBuilderInfrastructureConfiguration,
			},
			{
				Name:        "instance_types",
				Description: "The instance types of the infrastructure configuration.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "logging",
				Description: "The logging configuration of the infrastructure configuration.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getEc2ImageBuilderInfrastructureConfiguration,
			},
			{
				Name:        "placement",
				Description: "The instance placement settings that define where the instances that are launched from your image run.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "resource_tags",
				Description: "The tags attached to the resource created by Image Builder.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "security_group_ids",
				Description: "The security group IDs of the infrastructure configuration.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getEc2ImageBuilderInfrastructureConfiguration,
			},
			{
				Name:        "tags_src",
				Description: "A map of tags attached to the infrastructure configuration.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},

			// Standard columns
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "tags",
				Description: resourceInterfaceDescription("tags"),
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Tags"),
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Arn").Transform(arnToAkas),
			},
		}),
	}
}

func listEc2ImageBuilderInfrastructureConfigurations(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	svc, err := ImageBuilderClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_ec2_image_builder_infrastructure_configuration.listEc2ImageBuilderInfrastructureConfigurations", "connection_error", err)
		return nil, err
	}

	maxItems := int32(25)
	if d.QueryContext.Limit != nil {
		limit := int32(*d.QueryContext.Limit)
		if limit < maxItems {
			if limit < 1 {
				maxItems = 1
			} else {
				maxItems = limit
			}
		}
	}

	input := &imagebuilder.ListInfrastructureConfigurationsInput{
		MaxResults: aws.Int32(maxItems),
	}
	paginator := imagebuilder.NewListInfrastructureConfigurationsPaginator(svc, input, func(o *imagebuilder.ListInfrastructureConfigurationsPaginatorOptions) {
		o.Limit = maxItems
		o.StopOnDuplicateToken = true
	})

	for paginator.HasMorePages() {
		d.WaitForListRateLimit(ctx)

		output, err := paginator.NextPage(ctx)
		if err != nil {
			plugin.Logger(ctx).Error("aws_ec2_image_builder_infrastructure_configuration.listEc2ImageBuilderInfrastructureConfigurations", "api_error", err)
			return nil, err
		}

		for _, config := range output.InfrastructureConfigurationSummaryList {
			d.StreamListItem(ctx, config)

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}

	return nil, nil
}

func getEc2ImageBuilderInfrastructureConfiguration(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	arn := d.EqualsQualString("arn")
	if h.Item != nil {
		arn = ec2ImageBuilderInfrastructureConfigurationArn(h.Item)
	}
	if arn == "" {
		return nil, nil
	}

	svc, err := ImageBuilderClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_ec2_image_builder_infrastructure_configuration.getEc2ImageBuilderInfrastructureConfiguration", "connection_error", err)
		return nil, err
	}

	input := &imagebuilder.GetInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(arn),
	}
	output, err := svc.GetInfrastructureConfiguration(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("aws_ec2_image_builder_infrastructure_configuration.getEc2ImageBuilderInfrastructureConfiguration", "api_error", err)
		return nil, err
	}

	if output.InfrastructureConfiguration == nil {
		return nil, nil
	}
	return output.InfrastructureConfiguration, nil
}

func ec2ImageBuilderInfrastructureConfigurationArn(item interface{}) string {
	switch item := item.(type) {
	case types.InfrastructureConfigurationSummary:
		return aws.ToString(item.Arn)
	case *types.InfrastructureConfiguration:
		return aws.ToString(item.Arn)
	}
	return ""
}
