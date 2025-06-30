package pluginaws

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/pluginaws/testutil"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

func TestSqsClient_PublishToQueue(t *testing.T) {
	blockCtx := types.NewBlockContext(1, time.Now())
	logger := log.NewLogger(os.Stdout)

	tests := []struct {
		name            string
		messages        []*types.Message
		mockResponses   []*sqs.SendMessageBatchOutput
		mockErrors      []error
		expectedError   bool
		expectedRetries int
	}{
		{
			name: "successful publish",
			messages: []*types.Message{
				types.NewMessage("test", "data1", blockCtx),
				types.NewMessage("test", "data2", blockCtx),
			},
			mockResponses: []*sqs.SendMessageBatchOutput{
				{
					Failed: []*sqs.BatchResultErrorEntry{},
				},
			},
			expectedError:   false,
			expectedRetries: 0,
		},
		{
			name: "partial failure with retry",
			messages: []*types.Message{
				types.NewMessage("test", "data1", blockCtx),
				types.NewMessage("test", "data2", blockCtx),
			},
			mockResponses: []*sqs.SendMessageBatchOutput{
				{
					Failed: []*sqs.BatchResultErrorEntry{
						{
							Id:          stringPtr("test-0"),
							Code:        stringPtr("InternalError"),
							Message:     stringPtr(""),
							SenderFault: boolPtr(false),
						},
					},
				},
				{
					Failed: []*sqs.BatchResultErrorEntry{},
				},
			},
			expectedError:   false,
			expectedRetries: 1,
		},
		{
			name: "multiple retries before success",
			messages: []*types.Message{
				types.NewMessage("test", "data1", blockCtx),
			},
			mockResponses: []*sqs.SendMessageBatchOutput{
				{
					Failed: []*sqs.BatchResultErrorEntry{
						{
							Id:          stringPtr("test-0"),
							Code:        stringPtr("InternalError"),
							Message:     stringPtr(""),
							SenderFault: boolPtr(false),
						},
					},
				},
				{
					Failed: []*sqs.BatchResultErrorEntry{
						{
							Id:          stringPtr("test-0"),
							Code:        stringPtr("InternalError"),
							Message:     stringPtr(""),
							SenderFault: boolPtr(false),
						},
					},
				},
				{
					Failed: []*sqs.BatchResultErrorEntry{},
				},
			},
			expectedError:   false,
			expectedRetries: 2,
		},
		{
			name: "sender fault - no retry",
			messages: []*types.Message{
				types.NewMessage("test", "data1", blockCtx),
			},
			mockResponses: []*sqs.SendMessageBatchOutput{
				{
					Failed: []*sqs.BatchResultErrorEntry{
						{
							Id:          stringPtr("test-0"),
							Code:        stringPtr("InvalidMessageContents"),
							Message:     stringPtr("Invalid message"),
							SenderFault: boolPtr(true),
						},
					},
				},
			},
			expectedError:   true,
			expectedRetries: 0,
		},
		{
			name: "max retries exceeded",
			messages: []*types.Message{
				types.NewMessage("test", "data1", blockCtx),
			},
			mockResponses: []*sqs.SendMessageBatchOutput{
				{
					Failed: []*sqs.BatchResultErrorEntry{
						{
							Id:          stringPtr("test-0"),
							Code:        stringPtr("InternalError"),
							Message:     stringPtr(""),
							SenderFault: boolPtr(false),
						},
					},
				},
				{
					Failed: []*sqs.BatchResultErrorEntry{
						{
							Id:          stringPtr("test-0"),
							Code:        stringPtr("InternalError"),
							Message:     stringPtr(""),
							SenderFault: boolPtr(false),
						},
					},
				},
				{
					Failed: []*sqs.BatchResultErrorEntry{
						{
							Id:          stringPtr("test-0"),
							Code:        stringPtr("InternalError"),
							Message:     stringPtr(""),
							SenderFault: boolPtr(false),
						},
					},
				},
			},
			expectedError:   true,
			expectedRetries: MaxAttempts,
		},
		{
			name: "batch size limit",
			messages: []*types.Message{
				types.NewMessage("test", "data1", blockCtx),
				types.NewMessage("test", "data2", blockCtx),
				types.NewMessage("test", "data3", blockCtx),
				types.NewMessage("test", "data4", blockCtx),
				types.NewMessage("test", "data5", blockCtx),
				types.NewMessage("test", "data6", blockCtx),
				types.NewMessage("test", "data7", blockCtx),
				types.NewMessage("test", "data8", blockCtx),
				types.NewMessage("test", "data9", blockCtx),
				types.NewMessage("test", "data10", blockCtx),
				types.NewMessage("test", "data11", blockCtx),
			},
			mockResponses: []*sqs.SendMessageBatchOutput{
				{
					Failed: []*sqs.BatchResultErrorEntry{},
				},
				{
					Failed: []*sqs.BatchResultErrorEntry{},
				},
			},
			expectedError:   false,
			expectedRetries: 0,
		},
		{
			name: "batch size limit with file",
			messages: []*types.Message{
				// 3 should normally fit in a single batch, but due to the large message body, it will be split into 2 batches
				// -100 is to account for the JSON marshalling overhead
				types.NewMessage("test", strings.Repeat("0", MaxMessageBodyLengthBytes-100), blockCtx),
				types.NewMessage("test", strings.Repeat("1", MaxMessageBodyLengthBytes-100), blockCtx),
				types.NewMessage("test", strings.Repeat("2", MaxMessageBodyLengthBytes-100), blockCtx),
			},
			mockResponses: []*sqs.SendMessageBatchOutput{
				{
					Failed: []*sqs.BatchResultErrorEntry{},
				},
				{
					Failed: []*sqs.BatchResultErrorEntry{},
				},
			},
			expectedError:   false,
			expectedRetries: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSQS := testutil.NewMockSQSAPI(ctrl)

			seenMessages := make(map[string]bool)
			for i, message := range tt.messages {
				seenMessages[fmt.Sprintf("%s-%d", message.Type, i)] = false
			}

			// Set up expectations for each call
			for i, response := range tt.mockResponses {
				var err error
				if tt.mockErrors != nil && i < len(tt.mockErrors) {
					err = tt.mockErrors[i]
				}

				mockSQS.EXPECT().
					SendMessageBatch(gomock.Any()).
					DoAndReturn(func(input *sqs.SendMessageBatchInput) (*sqs.SendMessageBatchOutput, error) {
						assert.Equal(t, "test-queue", *input.QueueUrl)
						assert.LessOrEqualf(t, len(input.Entries), MaxSQSBatchSize, "expected batch size to be less than or equal to %d, got %d", MaxSQSBatchSize, len(input.Entries))
						for _, entry := range input.Entries {
							seenMessages[*entry.Id] = true
						}
						return response, err
					})
			}

			client := &SqsClient{
				sqsClient: mockSQS,
				queueURL:  "test-queue",
				logger:    logger,
			}

			err := client.PublishToQueue(tt.messages)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			for id, seen := range seenMessages {
				assert.Truef(t, seen, "expected message %s to be seen", id)
			}
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

// regression test for bug encountered on testnet 30th June 2025
func TestTakeBatch_SizeCalculation(t *testing.T) {
	// Create entries with small message bodies but large attributes
	entries := []*sizedBatchEntry{
		{
			deliveryAttempts: 0,
			size:             50000,
			entry: &sqs.SendMessageBatchRequestEntry{
				Id:          stringPtr("test-0"),
				MessageBody: aws.String("small body"),
			},
		},
		{
			deliveryAttempts: 0,
			size:             50000,
			entry: &sqs.SendMessageBatchRequestEntry{
				Id:          stringPtr("test-1"),
				MessageBody: aws.String("small body"),
			},
		},
		{
			deliveryAttempts: 0,
			size:             50000,
			entry: &sqs.SendMessageBatchRequestEntry{
				Id:          stringPtr("test-2"),
				MessageBody: aws.String("small body"),
			},
		},
		{
			deliveryAttempts: 0,
			size:             50000,
			entry: &sqs.SendMessageBatchRequestEntry{
				Id:          stringPtr("test-3"),
				MessageBody: aws.String("small body"),
			},
		},
		{
			deliveryAttempts: 0,
			size:             50000,
			entry: &sqs.SendMessageBatchRequestEntry{
				Id:          stringPtr("test-4"),
				MessageBody: aws.String("small body"),
			},
		},
		{
			deliveryAttempts: 0,
			size:             50000,
			entry: &sqs.SendMessageBatchRequestEntry{
				Id:          stringPtr("test-5"),
				MessageBody: aws.String("small body"),
			},
		},
	}

	// With the buggy implementation, takeBatch would include all 6 entries because
	// it only looks at message body size and not the attributes
	batch := takeBatch(entries)

	// Calculate the actual size of the batch
	totalSize := 0
	for _, entry := range batch {
		totalSize += entry.size
	}

	assert.LessOrEqual(t, totalSize, MaxAwsRequestLengthBytes,
		"Batch total size should not exceed %d bytes, got %d", MaxAwsRequestLengthBytes, totalSize)
}
