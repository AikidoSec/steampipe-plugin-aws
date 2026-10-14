package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableAwsAPIGatewayVpcLink(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_api_gateway_vpc_link",
		Description: "AWS API Gateway VPC Link",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: shouldIgnoreErrors([]string{"NotFoundException"}),
			},
			Hydrate: getAPIGatewayVpcLink,
			Tags:    map[string]string{"service": "apigateway", "action": "GetVpcLink"},
		},
		List: &plugin.ListConfig{
			Hydrate: listAPIGatewayVpcLinks,
			Tags:    map[string]string{"service": "apigateway", "action": "GetVpcLinks"},
		},
		GetMatrixItemFunc: SupportedRegionMatrix(AWS_APIGATEWAY_SERVICE_ID),
		Columns: awsRegionalColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The identifier of the VpcLink. It is used in an Integration to reference this VpcLink",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The name used to label and identify the VPC link",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "The description of the VPC link",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The status of the VPC link. The valid values are AVAILABLE, PENDING, DELETING, or FAILED",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status_message",
				Description: "A description about the VPC link status",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "target_arns",
				Description: "The ARN of the network load balancer of the VPC targeted by the VPC link. The network load balancer must be owned by the same AWS account of the API owner",
				Type:        proto.ColumnType_JSON,
			},

			//// Steampipe standard columns
			{
				Name:        "tags",
				Description: resourceInterfaceDescription("tags"),
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Hydrate:     getAPIGatewayVpcLinkAkas,
				Transform:   transform.FromValue(),
			},
		}),
	}
}

//// LIST FUNCTION

func listAPIGatewayVpcLinks(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)

	// Create service
	svc, err := APIGatewayClient(ctx, d)
	if err != nil {
		logger.Error("aws_api_gateway_vpc_link.listAPIGatewayVpcLinks", "connection error", err)
		return nil, err
	}

	// Limiting the results
	maxLimit := int32(500)
	if d.QueryContext.Limit != nil {
		limit := int32(*d.QueryContext.Limit)
		if limit < maxLimit {
			if limit < 1 {
				maxLimit = 1
			} else {
				maxLimit = limit
			}
		}
	}

	input := &apigateway.GetVpcLinksInput{
		Limit: aws.Int32(maxLimit),
	}

	paginator := apigateway.NewGetVpcLinksPaginator(svc, input, func(o *apigateway.GetVpcLinksPaginatorOptions) {
		o.Limit = maxLimit
		o.StopOnDuplicateToken = true
	})

	// List call
	for paginator.HasMorePages() {
		// apply rate limiting
		d.WaitForListRateLimit(ctx)

		output, err := paginator.NextPage(ctx)
		if err != nil {
			logger.Error("aws_api_gateway_vpc_link.listAPIGatewayVpcLinks", "api_error", err)
			return nil, err
		}

		for _, vpcLink := range output.Items {
			d.StreamListItem(ctx, vpcLink)

			// Context can be cancelled due to manual cancellation or the limit has been hit
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getAPIGatewayVpcLink(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {

	// Create service
	svc, err := APIGatewayClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_api_gateway_vpc_link.getAPIGatewayVpcLink", "connection_error", err)
		return nil, err
	}

	id := d.EqualsQuals["id"].GetStringValue()

	// Empty check
	if id == "" {
		return nil, nil
	}

	params := &apigateway.GetVpcLinkInput{
		VpcLinkId: aws.String(id),
	}

	vpcLink, err := svc.GetVpcLink(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "NotFoundException") {
			return nil, nil
		}
		plugin.Logger(ctx).Error("aws_api_gateway_vpc_link.getAPIGatewayVpcLink", "api_error", err)
		return nil, err
	}

	return vpcLink, nil
}

func getAPIGatewayVpcLinkAkas(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	id := ""

	switch item := h.Item.(type) {
	case *apigateway.GetVpcLinkOutput:
		id = *item.Id
	case types.VpcLink:
		id = *item.Id
	}

	commonData, err := getCommonColumns(ctx, d, h)
	if err != nil {
		plugin.Logger(ctx).Error("aws_api_gateway_vpc_link.getAPIGatewayVpcLinkAkas", "getCommonColumns_error", err)
		return nil, err
	}

	commonColumnData := commonData.(*awsCommonColumnData)
	akas := []string{"arn:" + commonColumnData.Partition + ":apigateway:" + region + "::/vpclinks/" + id}

	return akas, nil
}
