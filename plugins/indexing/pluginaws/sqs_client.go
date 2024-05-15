package pluginaws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

const (
	MaxAwsRequestLengthBytes = 262_144 // 256KB
	MaxSQSBatchSize          = 10
)

func (sc *SqsClient) PublishToQueue(data []*types.Message) error {
	sc.logger.Trace("publishing to queue", "size", len(data))

	allowedMessages := sc.filterMessages(data)
	sizedBatchEntries, err := sc.createSizedBatchEntries(allowedMessages)
	if err != nil {
		return err
	}

	currentRequestSize := 0
	entries := make([]*sqs.SendMessageBatchRequestEntry, 0, MaxSQSBatchSize)
	for _, sizedBatchEntry := range sizedBatchEntries {
		nextSize := currentRequestSize + sizedBatchEntry.size

		if len(entries) == MaxSQSBatchSize || nextSize > MaxAwsRequestLengthBytes {
			if err := sc.sendMessageBatch(entries); err != nil {
				return err
			}

			entries = nil
			currentRequestSize = 0
		}

		entries = append(entries, sizedBatchEntry.entry)
		currentRequestSize += sizedBatchEntry.size
	}

	if len(entries) > 0 {
		if err := sc.sendMessageBatch(entries); err != nil {
			return err
		}
	}

	return nil
}

func (sc *SqsClient) sendMessageBatch(batch []*sqs.SendMessageBatchRequestEntry) error {
	sc.logger.Trace("sending batch", "size", len(batch))

	result, err := sc.sqsClient.SendMessageBatch(&sqs.SendMessageBatchInput{
		Entries:  batch,
		QueueUrl: &sc.queueURL,
	})
	if err != nil {
		return err
	}

	sc.logger.Trace("send batch request succeeded")

	if len(result.Failed) > 0 {
		return fmt.Errorf("failed to send %d messages", len(result.Failed))
	}

	return nil
}
