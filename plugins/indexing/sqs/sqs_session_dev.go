//go:build dev

package pluginsqs

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

var endpointEnvName = "SQS_ENDPOINT"

func NewSession() (*session.Session, error) {
	endpoint, found := os.LookupEnv(endpointEnvName)
	if !found {
		panic(fmt.Errorf("missing environment variable '%s'", endpointEnvName))
	}

	return session.NewSession(&aws.Config{
		Region:      aws.String("eu-west-2"),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
		Endpoint:    aws.String(endpoint),
	})
}
