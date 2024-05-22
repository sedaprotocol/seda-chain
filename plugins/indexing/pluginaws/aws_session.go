//go:build !dev

package pluginaws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
)

func NewSession() (*session.Session, error) {
	return session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
}

func NewS3Config() (*aws.Config, error) {
	cfg := aws.NewConfig()
	request.WithRetryer(cfg, CustomRetryer{DefaultRetryer: client.DefaultRetryer{
		NumMaxRetries:    client.DefaultRetryerMaxNumRetries,
		MinRetryDelay:    client.DefaultRetryerMinRetryDelay,
		MaxRetryDelay:    client.DefaultRetryerMaxRetryDelay,
		MinThrottleDelay: client.DefaultRetryerMinThrottleDelay,
		MaxThrottleDelay: client.DefaultRetryerMaxThrottleDelay,
	}})
	return cfg, nil
}
