package pluginaws

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
)

var (
	queueURLEnvName      = "SQS_QUEUE_URL"
	bucketURLEnvName     = "S3_LARGE_MSG_BUCKET_NAME"
	messageFilterEnvName = "ALLOWED_MESSAGE_TYPES"
)

type SqsClient struct {
	bucketName      string
	queueURL        string
	allowedMessages []string
	sqsClient       *sqs.SQS
	s3Client        *s3.S3
	logger          *log.Logger
}

func NewSqsClient(logger *log.Logger) (*SqsClient, error) {
	queueURL, found := os.LookupEnv(queueURLEnvName)
	if !found {
		return nil, fmt.Errorf("missing environment variable '%s'", queueURLEnvName)
	}

	bucketName, found := os.LookupEnv(bucketURLEnvName)
	if !found {
		return nil, fmt.Errorf("missing environment variable '%s'", bucketURLEnvName)
	}

	allowedMessages := make([]string, 0)
	messageFilterString, found := os.LookupEnv(messageFilterEnvName)
	if found && messageFilterString != "" {
		allowedMessages = parseMessageFilterString(messageFilterString)
	}

	sess, err := NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to initialise session: %w", err)
	}
	sqsConfig := aws.NewConfig()
	AddRetryToConfig(sqsConfig)
	awsSqsClient := sqs.New(sess, sqsConfig)

	s3Config, err := NewS3Config()
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 config: %w", err)
	}
	awsS3Client := s3.New(sess, s3Config)

	return &SqsClient{
		queueURL:        queueURL,
		bucketName:      bucketName,
		allowedMessages: allowedMessages,
		sqsClient:       awsSqsClient,
		s3Client:        awsS3Client,
		logger:          logger,
	}, nil
}

func parseMessageFilterString(data string) []string {
	return strings.Split(data, ",")
}
