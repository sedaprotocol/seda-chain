package pluginaws

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	snsiface "github.com/aws/aws-sdk-go/service/sns/snsiface"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
)

var (
	topicARNEnvName      = "SNS_TOPIC_ARN"
	bucketURLEnvName     = "S3_LARGE_MSG_BUCKET_NAME"
	messageFilterEnvName = "ALLOWED_MESSAGE_TYPES"
)

type SnsClient struct {
	bucketName      string
	topicARN        string
	allowedMessages []string
	snsClient       snsiface.SNSAPI
	s3Client        *s3.S3
	logger          *log.Logger
}

func NewSnsClient(logger *log.Logger) (*SnsClient, error) {
	queueURL, found := os.LookupEnv(topicARNEnvName)
	if !found {
		return nil, fmt.Errorf("missing environment variable '%s'", topicARNEnvName)
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
	snsConfig := aws.NewConfig()
	AddRetryToConfig(snsConfig)
	awsSnsClient := sns.New(sess, snsConfig)

	s3Config, err := NewS3Config()
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 config: %w", err)
	}
	awsS3Client := s3.New(sess, s3Config)

	return &SnsClient{
		topicARN:        queueURL,
		bucketName:      bucketName,
		allowedMessages: allowedMessages,
		snsClient:       awsSnsClient,
		s3Client:        awsS3Client,
		logger:          logger,
	}, nil
}

func parseMessageFilterString(data string) []string {
	return strings.Split(data, ",")
}
