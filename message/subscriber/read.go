package subscriber

import (
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

func BulkRead(messagesCh <-chan *message.Message, limit int, timeout time.Duration) (receivedMessages message.Messages, all bool) {
	timeouted := time.After(timeout)

	// todo - copy move down

MessagesLoop:
	for len(receivedMessages) < limit {
		select {
		case msg, ok := <-messagesCh:
			if !ok {
				break MessagesLoop
			}

			receivedMessages = append(receivedMessages, msg)
			msg.Ack()
		case <-timeouted:
			break MessagesLoop
		}
	}

	return receivedMessages, len(receivedMessages) == limit
}

// todo -add tests & deduplicate
func BulkReadWithDeduplication(messagesCh <-chan *message.Message, limit int, timeout time.Duration) (receivedMessages message.Messages, all bool) {
	allMessagesReceived := make(chan struct{}, 1)

	receivedIDs := map[string]struct{}{}

	go func() {
		for msg := range messagesCh {
			if _, alreadyReceived := receivedIDs[msg.UUID]; !alreadyReceived {
				receivedMessages = append(receivedMessages, msg)
				receivedIDs[msg.UUID] = struct{}{}
			}
			msg.Ack()

			if len(receivedMessages) == limit {
				allMessagesReceived <- struct{}{}
				break
			}
		}
		// messagesCh closed
		allMessagesReceived <- struct{}{}
	}()

	select {
	case <-allMessagesReceived:
	case <-time.After(timeout):
	}

	return receivedMessages, len(receivedMessages) == limit
}
