package pluginsqs

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"

	types "github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

var queueURLEnvName = "SQS_QUEUE_URL"

type SqsClient struct {
	queueURL  string
	sqsClient *sqs.SQS
}

func (sc *SqsClient) sendMessageBatch(batch []*sqs.SendMessageBatchRequestEntry) ([]*sqs.BatchResultErrorEntry, error) {
	result, err := sc.sqsClient.SendMessageBatch(&sqs.SendMessageBatchInput{
		Entries:  batch,
		QueueUrl: &sc.queueURL,
	})
	if err != nil {
		return nil, err
	}

	return result.Failed, nil
}

func (sc *SqsClient) PublishToQueue(data []*types.Message) error {
	// Remember max message size is 262,144 bytes
	entries := make([]*sqs.SendMessageBatchRequestEntry, 0, 10)

	for i, message := range data {
		serialisedMessage, err := json.Marshal(message)
		if err != nil {
			return err
		}

		entries = append(entries, &sqs.SendMessageBatchRequestEntry{
			Id: aws.String(fmt.Sprintf("%s-%d", message.Type, i)),
			// TODO Create an enum for different types (based on foreign key constraints at the other end)
			MessageGroupId: aws.String("chain_events"),
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"height": {
					DataType:    aws.String("Number"),
					StringValue: aws.String(strconv.FormatInt(message.Block.Height, 10)),
				},
				"time": {
					DataType:    aws.String("String"),
					StringValue: aws.String(message.Block.Time.Format(time.RFC3339)),
				},
			},
			MessageBody: aws.String(string(serialisedMessage)),
		})

		if len(entries) == 10 {
			failed, err := sc.sendMessageBatch(entries)
			if err != nil {
				return err
			}

			if len(failed) > 0 {
				return fmt.Errorf("failed to send %d messages", len(failed))
			}

			entries = nil
		}
	}

	if len(entries) > 0 {
		failed, err := sc.sendMessageBatch(entries)
		if err != nil {
			return err
		}

		if len(failed) > 0 {
			return fmt.Errorf("failed to send %d messages", len(failed))
		}
	}

	return nil
}

func NewSqsClient() (*SqsClient, error) {
	queueURL, found := os.LookupEnv(queueURLEnvName)
	if !found {
		return nil, fmt.Errorf("missing environment variable '%s'", queueURLEnvName)
	}

	sess, err := NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to initialise session: %w", err)
	}

	awsSqsClient := sqs.New(sess)

	return &SqsClient{
		sqsClient: awsSqsClient,
		queueURL:  queueURL,
	}, nil
}
