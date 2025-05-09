package pluginaws

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

const (
	// Roughly half the limit of SQS request size (256KiB)
	// since we're not expecting many messages to be that large
	// we want to leave room for other messages in the same batch
	MaxMessageBodyLengthBytes = 120_000 // ~120KB
)

func (sc *SqsClient) uploadToS3(key string, body []byte, ctx *types.BlockContext) (*types.Message, error) {
	sc.logger.Trace("uploading to S3", "key", key)

	response, err := sc.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(sc.bucketName),
		Body:   bytes.NewReader(body),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	sc.logger.Trace("upload to S3 successful", "key", key, "etag", *response.ETag)

	data := struct {
		Key  string `json:"key"`
		ETag string `json:"ETag"`
	}{
		Key:  key,
		ETag: *response.ETag,
	}
	return types.NewMessage("large-message", data, ctx), nil
}

func batchEntrySize(msg *string, msgAttr map[string]*sqs.MessageAttributeValue) (int, error) {
	var size int
	if msg != nil {
		size += len(*msg)
	}

	for k, v := range msgAttr {
		dataType := v.DataType
		size += len(k)
		size += len(*dataType)
		switch {
		case strings.HasPrefix(*dataType, "String") || strings.HasPrefix(*dataType, "Number"):
			size += len(*v.StringValue)
		case strings.HasPrefix(*dataType, "Binary"):
			size += len(v.BinaryValue)
		default:
			return -1, fmt.Errorf("unexpected data type: %s", *dataType)
		}
	}

	return size, nil
}
