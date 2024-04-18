//go:build dev

package pluginaws

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	sqsEndpointEnvName = "SQS_ENDPOINT"
	s3EndpointEnvName  = "S3_ENDPOINT"
)

func NewSession() (*session.Session, error) {
	endpoint, found := os.LookupEnv(sqsEndpointEnvName)
	if !found {
		return nil, fmt.Errorf("missing environment variable '%s'", sqsEndpointEnvName)
	}

	return session.NewSession(&aws.Config{
		Region:      aws.String("eu-west-2"),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
		Endpoint:    aws.String(endpoint),
	})
}

func NewS3Config() (*aws.Config, error) {
	endpoint, found := os.LookupEnv(s3EndpointEnvName)
	if !found {
		return nil, fmt.Errorf("missing environment variable '%s'", s3EndpointEnvName)
	}
	// The local emulator requires path style access
	return aws.NewConfig().WithEndpoint(endpoint).WithS3ForcePathStyle(true), nil
}
