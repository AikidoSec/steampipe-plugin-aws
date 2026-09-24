package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func TestGetEcsServiceAutoscalingResourceID(t *testing.T) {
	for _, tc := range []struct {
		name    string
		service types.Service
		want    interface{}
	}{
		{"short ARN", types.Service{ClusterArn: aws.String("arn:aws:ecs:eu-west-1:123456789012:cluster/production"), ServiceName: aws.String("api"), ServiceArn: aws.String("arn:aws:ecs:eu-west-1:123456789012:service/api")}, "service/production/api"},
		{"long ARN with service suffix in cluster", types.Service{ClusterArn: aws.String("arn:aws:ecs:eu-west-1:123456789012:cluster/myservice"), ServiceName: aws.String("api"), ServiceArn: aws.String("arn:aws:ecs:eu-west-1:123456789012:service/myservice/api")}, "service/myservice/api"},
		{"missing cluster", types.Service{ServiceName: aws.String("api")}, nil},
		{"missing service name", types.Service{ClusterArn: aws.String("arn:aws:ecs:eu-west-1:123456789012:cluster/production")}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getEcsServiceAutoscalingResourceID(context.Background(), &transform.TransformData{HydrateItem: tc.service})
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
