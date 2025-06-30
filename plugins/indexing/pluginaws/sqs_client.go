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

	for len(sizedBatchEntries) > 0 {
		batch := takeBatch(sizedBatchEntries)

		failedEntries, err := sc.sendMessageBatch(batch)
		if err != nil {
			return err
		}

		// Append failed entries to the back of the list for retry
		if failedEntries != nil {
			sizedBatchEntries = append(sizedBatchEntries, failedEntries...)
		}

		// Remove processed entries from the front of the list
		sizedBatchEntries = sizedBatchEntries[len(batch):]
	}

	return nil
}

// takeBatch takes a list of retryable entries and returns a batch of entries from the front of the list
// that fit in a single SQS request.
func takeBatch(retryableEntries []*sizedBatchEntry) []*sizedBatchEntry {
	currentRequestSize := 0
	batch := make([]*sizedBatchEntry, 0, MaxSQSBatchSize)

	for _, retryableEntry := range retryableEntries {
		nextSize := currentRequestSize + retryableEntry.size
		if len(batch) == MaxSQSBatchSize || nextSize > MaxAwsRequestLengthBytes {
			return batch
		}

		batch = append(batch, retryableEntry)
		currentRequestSize += retryableEntry.size
	}

	return batch
}

// sendMessageBatch attempts to send a batch of messages to SQS.
// It returns a list of failed entries that can be retried.
// If any of the messages are not retryable, exceed the max retry attempts,
// or the request failed after the SDK's retry attempts, the function will return an error.
func (sc *SqsClient) sendMessageBatch(entries []*sizedBatchEntry) ([]*sizedBatchEntry, error) {
	sc.logger.Trace("sending batch", "size", len(entries))

	// Prepare a message batch and increment the attempts for each entry
	batch := make([]*sqs.SendMessageBatchRequestEntry, 0, len(entries))
	for _, entry := range entries {
		batch = append(batch, entry.entry)
		entry.incrementAttempts()
	}

	// Send the message batch, request level retry is handled by the SDK.
	result, err := sc.sqsClient.SendMessageBatch(&sqs.SendMessageBatchInput{
		Entries:  batch,
		QueueUrl: &sc.queueURL,
	})
	if err != nil {
		sc.logger.Error("failed to send message batch", "error", err.Error())
		return nil, err
	}

	sc.logger.Trace("send batch request succeeded")

	// It's possible that the request succeeded but some messages failed to be delivered,
	// those are returned in the result.Failed slice.
	if len(result.Failed) == 0 {
		return nil, nil
	}

	sc.logger.Trace("send batch request succeeded with failed entries", "size", len(result.Failed))

	failedEntries := make([]*sizedBatchEntry, 0, len(result.Failed))
	for _, failed := range result.Failed {
		// As all the values are pointers, we need to check if they are nil
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

		sc.logger.Error("failed to deliver message", "error", errorMessage, "id", id, "code", errorCode, "senderFault", senderFault)

		// If the message didn't fail due to a sender fault, we should be able to retry it.
		if senderFault == "false" {
			entry := findBatchRequestEntry(entries, id)

			if entry == nil {
				return nil, fmt.Errorf("failed to find entry %s", id)
			}

			if entry.attemptsExceeded() {
				return nil, fmt.Errorf("failed to deliver message %s after %d attempts", id, MaxAttempts)
			}

			failedEntries = append(failedEntries, entry)
		}
	}

	// Failed entries can be retried so we do not return an error.
	if len(failedEntries) > 0 {
		sc.logger.Trace("retrying failed entries", "size", len(failedEntries))
		return failedEntries, nil
	}

	// If we received result.Failed entries but none of them were retryable, we should
	// stop the indexing process.
	return nil, fmt.Errorf("failed to deliver %d messages", len(failedEntries))
}

// findBatchRequestEntry finds an entry in the batch by id.
func findBatchRequestEntry(entries []*sizedBatchEntry, id string) *sizedBatchEntry {
	for _, entry := range entries {
		if entry.entry.Id != nil && *entry.entry.Id == id {
			return entry
		}
	}

	return nil
}
