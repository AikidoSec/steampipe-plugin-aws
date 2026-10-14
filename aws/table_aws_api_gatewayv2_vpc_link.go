package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableAwsAPIGatewayV2VpcLink(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_api_gatewayv2_vpc_link",
		Description: "AWS API Gateway Version 2 VPC Link",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("vpc_link_id"),
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: shouldIgnoreErrors([]string{"NotFoundException"}),
			},
			Hydrate: getAPIGatewayV2VpcLink,
			Tags:    map[string]string{"service": "apigateway", "action": "GetVpcLink"},
		},
		List: &plugin.ListConfig{
			Hydrate: listAPIGatewayV2VpcLinks,
			Tags:    map[string]string{"service": "apigateway", "action": "GetVpcLinks"},
		},
		GetMatrixItemFunc: SupportedRegionMatrix(AWS_APIGATEWAY_SERVICE_ID),
		Columns: awsRegionalColumns([]*plugin.Column{
			{
				Name:        "vpc_link_id",
				Description: "The ID of the VPC link",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The name of the VPC link",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "vpc_link_status",
				Description: "The status of the VPC link",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "vpc_link_status_message",
				Description: "A message summarizing the cause of the status of the VPC link",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "vpc_link_version",
				Description: "The version of the VPC link",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "created_date",
				Description: "The timestamp when the VPC link was created",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "subnet_ids",
				Description: "A list of subnet IDs to include in the VPC link",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "security_group_ids",
				Description: "A list of security group IDs for the VPC link",
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
				Hydrate:     getAPIGatewayV2VpcLinkAkas,
				Transform:   transform.FromValue(),
			},
		}),
	}
}

//// LIST FUNCTION

func listAPIGatewayV2VpcLinks(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)

	// Create Session
	svc, err := APIGatewayV2Client(ctx, d)
	if err != nil {
		logger.Error("aws_api_gatewayv2_vpc_link.listAPIGatewayV2VpcLinks", "connection error", err)
		return nil, err
	}
	if svc == nil {
		// Unsupported region, return no data
		return nil, nil
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

	pagesLeft := true
	params := &apigatewayv2.GetVpcLinksInput{
		MaxResults: aws.String(fmt.Sprint(maxLimit)),
	}

	for pagesLeft {
		// apply rate limiting
		d.WaitForListRateLimit(ctx)

		result, err := svc.GetVpcLinks(ctx, params)
		if err != nil {
			logger.Error("aws_api_gatewayv2_vpc_link.listAPIGatewayV2VpcLinks", "api_error", err)
			return nil, err
		}

		for _, vpcLink := range result.Items {
			d.StreamListItem(ctx, vpcLink)

			// Context can be cancelled due to manual cancellation or the limit has been hit
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}

		if result.NextToken != nil {
			pagesLeft = true
			params.NextToken = result.NextToken
		} else {
			pagesLeft = false
		}
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getAPIGatewayV2VpcLink(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {

	// Create Session
	svc, err := APIGatewayV2Client(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_api_gatewayv2_vpc_link.getAPIGatewayV2VpcLink", "service_client_error", err)
		return nil, err
	}
	if svc == nil {
		// Unsupported region, return no data
		return nil, nil
	}

	id := d.EqualsQuals["vpc_link_id"].GetStringValue()

	// Empty check
	if id == "" {
		return nil, nil
	}

	params := &apigatewayv2.GetVpcLinkInput{
		VpcLinkId: aws.String(id),
	}

	vpcLink, err := svc.GetVpcLink(ctx, params)
	if err != nil {
		plugin.Logger(ctx).Error("aws_api_gatewayv2_vpc_link.getAPIGatewayV2VpcLink", "api_error", err)
		return nil, err
	}

	return vpcLink, nil
}

func getAPIGatewayV2VpcLinkAkas(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	region := d.EqualsQualString(matrixKeyRegion)
	id := ""

	switch item := h.Item.(type) {
	case *apigatewayv2.GetVpcLinkOutput:
		id = *item.VpcLinkId
	case types.VpcLink:
		id = *item.VpcLinkId
	}

	commonData, err := getCommonColumns(ctx, d, h)
	if err != nil {
		plugin.Logger(ctx).Error("aws_api_gatewayv2_vpc_link.getAPIGatewayV2VpcLinkAkas", "getCommonColumns_error", err)
		return nil, err
	}

	commonColumnData := commonData.(*awsCommonColumnData)
	akas := []string{"arn:" + commonColumnData.Partition + ":apigateway:" + region + "::/vpclinks/" + id}

	return akas, nil
}
