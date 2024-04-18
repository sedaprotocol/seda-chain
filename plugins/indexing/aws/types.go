package pluginaws

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"

	log "github.com/sedaprotocol/seda-chain/plugins/indexing/log"
)

var (
	queueURLEnvName  = "SQS_QUEUE_URL"
	bucketURLEnvName = "S3_LARGE_MSG_BUCKET_NAME"
)

type SqsClient struct {
	bucketName string
	queueURL   string
	sqsClient  *sqs.SQS
	s3Client   *s3.S3
	logger     *log.Logger
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

	sess, err := NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to initialise session: %w", err)
	}
	awsSqsClient := sqs.New(sess)

	s3Config, err := NewS3Config()
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 config: %w", err)
	}
	awsS3Client := s3.New(sess, s3Config)

	return &SqsClient{
		queueURL:   queueURL,
		bucketName: bucketName,
		sqsClient:  awsSqsClient,
		s3Client:   awsS3Client,
		logger:     logger,
	}, nil
}
