package pluginaws

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

const MaxAttempts = 3

type sizedBatchEntry struct {
	deliveryAttempts int
	size             int
	entry            *sns.PublishBatchRequestEntry
}

func (r *sizedBatchEntry) attemptsExceeded() bool {
	return r.deliveryAttempts >= MaxAttempts
}

func (r *sizedBatchEntry) incrementAttempts() {
	r.deliveryAttempts++
}

func newSizedBatchEntry(entry *sns.PublishBatchRequestEntry) (*sizedBatchEntry, error) {
	size, err := batchEntrySize(entry.Message, entry.MessageAttributes)
	if err != nil {
		return nil, err
	}

	return &sizedBatchEntry{
		deliveryAttempts: 0,
		size:             size,
		entry:            entry,
	}, nil
}

func (sc *SnsClient) createSizedBatchEntries(data []*types.Message) ([]*sizedBatchEntry, error) {
	entries := make([]*sizedBatchEntry, 0, len(data))

	for i, message := range data {
		serialisedMessage, err := json.Marshal(message)
		if err != nil {
			return nil, err
		}

		attributes := map[string]*sns.MessageAttributeValue{
			"height": {
				DataType:    aws.String("Number"),
				StringValue: aws.String(strconv.FormatInt(message.Block.Height, 10)),
			},
			"time": {
				DataType:    aws.String("String"),
				StringValue: aws.String(message.Block.Time.Format(time.RFC3339)),
			},
		}

		if len(serialisedMessage) > MaxMessageBodyLengthBytes {
			fileKey := fmt.Sprintf("%s-h%d-i%d.json", message.Type, message.Block.Height, i)
			s3Message, err := sc.uploadToS3(fileKey, serialisedMessage, message.Block)
			if err != nil {
				return nil, err
			}

			s3SerialisedMessage, err := json.Marshal(s3Message)
			if err != nil {
				return nil, err
			}

			serialisedMessage = s3SerialisedMessage
			attributes["file"] = &sns.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(fileKey),
			}
		}

		sizedEntry, err := newSizedBatchEntry(&sns.PublishBatchRequestEntry{
			Id:                aws.String(fmt.Sprintf("%s-%d", message.Type, i)),
			MessageAttributes: attributes,
			Message:           aws.String(string(serialisedMessage)),
		})
		if err != nil {
			return nil, err
		}

		entries = append(entries, sizedEntry)
	}

	return entries, nil
}
