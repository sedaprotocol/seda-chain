package pluginaws

import (
	"slices"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

func (sc *SqsClient) filterMessages(data []*types.Message) []*types.Message {
	allowedMessages := make([]*types.Message, 0, len(data))

	for _, message := range data {
		if sc.isMessageAllowed(message) {
			allowedMessages = append(allowedMessages, message)
		} else {
			sc.logger.Trace("skipping message", "type", message.Type)
		}
	}

	return allowedMessages
}

func (sc *SqsClient) isMessageAllowed(event *types.Message) bool {
	// When no allowlist is specified assume everything is allowed.
	if len(sc.allowedMessages) == 0 {
		return true
	}

	return slices.Contains(sc.allowedMessages, event.Type)
}
