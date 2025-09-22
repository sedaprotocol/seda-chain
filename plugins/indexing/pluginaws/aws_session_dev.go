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
	snsEndpointEnvName = "SNS_ENDPOINT"
	s3EndpointEnvName  = "S3_ENDPOINT"
)

func NewSession() (*session.Session, error) {
	endpoint, found := os.LookupEnv(snsEndpointEnvName)
	if !found {
		return nil, fmt.Errorf("missing environment variable '%s'", snsEndpointEnvName)
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
	cfg := aws.NewConfig().WithEndpoint(endpoint).WithS3ForcePathStyle(true)
	AddRetryToConfig(cfg)
	return cfg, nil
}
