package pluginaws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

const (
	MaxAwsRequestLengthBytes = 262_144 // 256KiB
	MaxSQSBatchSize          = 10
	NilString                = "nil"
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
		for _, failed := range result.Failed {
			errorMessage := NilString
			if failed.Message != nil {
				errorMessage = *failed.Message
			}

			id := NilString
			if failed.Id != nil {
				id = *failed.Id
			}

			errorCode := NilString
			if failed.Code != nil {
				errorCode = *failed.Code
			}

			senderFault := NilString
			if failed.SenderFault != nil {
				senderFault = strconv.FormatBool(*failed.SenderFault)
			}

			sc.logger.Error("failed to send message", "error", errorMessage, "id", id, "code", errorCode, "senderFault", senderFault)
		}
		return fmt.Errorf("failed to send %d messages", len(result.Failed))
	}

	return nil
}
