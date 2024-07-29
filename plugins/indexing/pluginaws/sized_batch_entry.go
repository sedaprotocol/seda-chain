package pluginaws

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

type sizedBatchEntry struct {
	size  int
	entry *sqs.SendMessageBatchRequestEntry
}

func newSizedBatchEntry(entry *sqs.SendMessageBatchRequestEntry) (*sizedBatchEntry, error) {
	size, err := batchEntrySize(entry.MessageBody, entry.MessageAttributes)
	if err != nil {
		return nil, err
	}

	return &sizedBatchEntry{
		size:  size,
		entry: entry,
	}, nil
}

func (sc *SqsClient) createSizedBatchEntries(data []*types.Message) ([]*sizedBatchEntry, error) {
	entries := make([]*sizedBatchEntry, 0, len(data))

	for i, message := range data {
		serialisedMessage, err := json.Marshal(message)
		if err != nil {
			return nil, err
		}

		attributes := map[string]*sqs.MessageAttributeValue{
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
			attributes["file"] = &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(fileKey),
			}
		}

		sizedEntry, err := newSizedBatchEntry(&sqs.SendMessageBatchRequestEntry{
			Id:                aws.String(fmt.Sprintf("%s-%d", message.Type, i)),
			MessageAttributes: attributes,
			MessageBody:       aws.String(string(serialisedMessage)),
		})
		if err != nil {
			return nil, err
		}

		entries = append(entries, sizedEntry)
	}

	return entries, nil
}
