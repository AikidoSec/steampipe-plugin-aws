package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/fms"
	"github.com/aws/aws-sdk-go-v2/service/fms/types"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableAwsFMSComplianceDetails(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_fms_compliance_details",
		Description: "AWS Firewall Manager Compliance Details",
		List: &plugin.ListConfig{
			KeyColumns: []*plugin.KeyColumn{
				{Name: "policy_id", Require: plugin.Required},
			},
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: shouldIgnoreErrors([]string{"InvalidInputException", "InvalidOperationException", "ResourceNotFoundException"}),
			},
			Hydrate: listFmsComplianceDetails,
			Tags:    map[string]string{"service": "fms", "action": "ListComplianceStatus"},
		},
		GetMatrixItemFunc: SupportedRegionMatrix(AWS_FMS_SERVICE_ID),
		Columns: awsRegionalColumns([]*plugin.Column{
			{
				Name:        "policy_id",
				Description: "The ID of the Firewall Manager policy.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "policy_owner",
				Description: "The AWS account that created the Firewall Manager policy.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "member_account",
				Description: "The AWS account that owns the evaluated resource.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "resource_id",
				Description: "The ID of the resource that is not compliant with the policy.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Violator.ResourceId"),
			},
			{
				Name:        "resource_type",
				Description: "The type of the resource that is not compliant with the policy.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Violator.ResourceType"),
			},
			{
				Name:        "violation_reason",
				Description: "The reason that the resource is not compliant with the policy.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Violator.ViolationReason"),
			},
			{
				Name:        "metadata",
				Description: "Metadata about the resource that does not comply with the policy scope.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Violator.Metadata"),
			},
			{
				Name:        "evaluation_limit_exceeded",
				Description: "Indicates if over 100 resources are noncompliant with the Firewall Manager policy.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "expired_at",
				Description: "A timestamp that indicates when the returned information should be considered out of date.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "issue_info_map",
				Description: "Details about problems with dependent services, such as WAF or Config, and the error message received that indicates the problem with the service.",
				Type:        proto.ColumnType_JSON,
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Violator.ResourceId"),
			},
		}),
	}
}

type FmsComplianceDetailInfo struct {
	EvaluationLimitExceeded bool
	ExpiredAt               *time.Time
	IssueInfoMap            map[string]string
	MemberAccount           *string
	PolicyId                *string
	PolicyOwner             *string
	Violator                types.ComplianceViolator
}

//// LIST FUNCTION

func listFmsComplianceDetails(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	policyID := d.EqualsQualString("policy_id")

	if policyID == "" {
		return nil, nil
	}

	svc, err := FMSClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_fms_compliance_details.listFmsComplianceDetails", "connection_error", err)
		return nil, err
	}

	if svc == nil {
		return nil, nil
	}

	maxItems := int32(100)
	if d.QueryContext.Limit != nil {
		limit := int32(*d.QueryContext.Limit)
		if limit < maxItems {
			maxItems = limit
		}
	}

	input := &fms.ListComplianceStatusInput{
		MaxResults: &maxItems,
		PolicyId:   &policyID,
	}

	paginator := fms.NewListComplianceStatusPaginator(svc, input, func(o *fms.ListComplianceStatusPaginatorOptions) {
		o.Limit = maxItems
		o.StopOnDuplicateToken = true
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			plugin.Logger(ctx).Error("aws_fms_compliance_details.listFmsComplianceDetails.listComplianceStatus", "api_error", err)
			return nil, err
		}

		for _, status := range output.PolicyComplianceStatusList {
			if status.MemberAccount == nil {
				continue
			}

			params := &fms.GetComplianceDetailInput{
				PolicyId:      &policyID,
				MemberAccount: status.MemberAccount,
			}

			op, err := svc.GetComplianceDetail(ctx, params)
			if err != nil {
				plugin.Logger(ctx).Error("aws_fms_compliance_details.listFmsComplianceDetails.getComplianceDetail", "api_error", err)
				return nil, err
			}

			if op == nil || op.PolicyComplianceDetail == nil {
				continue
			}

			detail := op.PolicyComplianceDetail
			for _, violator := range detail.Violators {
				d.StreamListItem(ctx, FmsComplianceDetailInfo{
					EvaluationLimitExceeded: detail.EvaluationLimitExceeded,
					ExpiredAt:               detail.ExpiredAt,
					IssueInfoMap:            detail.IssueInfoMap,
					MemberAccount:           detail.MemberAccount,
					PolicyId:                detail.PolicyId,
					PolicyOwner:             detail.PolicyOwner,
					Violator:                violator,
				})

				if d.RowsRemaining(ctx) == 0 {
					return nil, nil
				}
			}
		}
	}

	return nil, nil
}
