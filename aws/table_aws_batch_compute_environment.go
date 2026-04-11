package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAwsBatchComputeEnvironment(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_batch_compute_environment",
		Description: "AWS Batch Compute Environment",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("compute_environment_name"),
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: shouldIgnoreErrors([]string{"ComputeEnvironmentNotFoundException"}),
			},
			Hydrate: getBatchComputeEnvironment,
			Tags:    map[string]string{"service": "batch", "action": "DescribeComputeEnvironments"},
		},
		List: &plugin.ListConfig{
			Hydrate: listBatchComputeEnvironments,
			Tags:    map[string]string{"service": "batch", "action": "DescribeComputeEnvironments"},
		},
		GetMatrixItemFunc: SupportedRegionMatrix(AWS_BATCH_SERVICE_ID),
		Columns: awsRegionalColumns([]*plugin.Column{
			{
				Name:        "compute_environment_name",
				Description: "The name of the compute environment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "arn",
				Description: "The Amazon Resource Name (ARN) of the compute environment.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ComputeEnvironmentArn"),
			},
			{
				Name:        "ecs_cluster_arn",
				Description: "The Amazon ECS cluster ARN associated with the compute environment.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("EcsClusterArn"),
			},
			{
				Name:        "eks_configuration",
				Description: "The Amazon EKS configuration for the compute environment.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("EksConfiguration"),
			},
			{
				Name:        "compute_resources",
				Description: "The compute resources defined for the compute environment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "container_orchestration_type",
				Description: "The orchestration type of the compute environment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "context",
				Description: "Reserved.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "service_role",
				Description: "The IAM role that AWS Batch uses for the compute environment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "state",
				Description: "The state of the compute environment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The current status of the compute environment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status_reason",
				Description: "A short, human-readable string to provide additional details about the current status.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "type",
				Description: "The type of the compute environment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "unmanaged_v_cpus",
				Description: "The maximum number of vCPUs expected for an unmanaged compute environment.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("UnmanagedvCpus"),
			},
			{
				Name:        "update_policy",
				Description: "The update policy for the compute environment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "uuid",
				Description: "The UUID of the compute environment.",
				Type:        proto.ColumnType_STRING,
			},

			// Standard columns for all tables
			{
				Name:        "tags",
				Description: "The tags assigned to the compute environment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ComputeEnvironmentName"),
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ComputeEnvironmentArn").Transform(transform.EnsureStringArray),
			},
		}),
	}
}

func listBatchComputeEnvironments(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	svc, err := BatchClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_batch_compute_environment.listBatchComputeEnvironments", "client_error", err)
		return nil, err
	}

	if svc == nil {
		return nil, nil
	}

	maxLimit := int32(100)
	if d.QueryContext.Limit != nil {
		limit := int32(*d.QueryContext.Limit)
		if limit < maxLimit {
			maxLimit = limit
		}
	}

	input := &batch.DescribeComputeEnvironmentsInput{
		MaxResults: aws.Int32(maxLimit),
	}

	paginator := batch.NewDescribeComputeEnvironmentsPaginator(svc, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			plugin.Logger(ctx).Error("aws_batch_compute_environment.listBatchComputeEnvironments", "api_error", err)
			return nil, err
		}

		for _, computeEnvironment := range output.ComputeEnvironments {
			d.StreamListItem(ctx, computeEnvironment)

			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}

	return nil, nil
}

func getBatchComputeEnvironment(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	computeEnvironmentName := d.EqualsQualString("compute_environment_name")
	if computeEnvironmentName == "" {
		return nil, nil
	}

	svc, err := BatchClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_batch_compute_environment.getBatchComputeEnvironment", "client_error", err)
		return nil, err
	}

	if svc == nil {
		return nil, nil
	}

	input := &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: []string{computeEnvironmentName},
	}

	output, err := svc.DescribeComputeEnvironments(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("aws_batch_compute_environment.getBatchComputeEnvironment", "api_error", err)
		return nil, err
	}

	if len(output.ComputeEnvironments) == 0 {
		return nil, nil
	}

	return output.ComputeEnvironments[0], nil
}
