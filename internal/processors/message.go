// Copyright 2019 Form3 Financial Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package processors

import (
	"context"
	"sync"

	"github.com/Altitude-sports/connector-sdk/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	log "github.com/sirupsen/logrus"

	"github.com/Altitude-sports/openfaas-sqs-connector/internal/pointers"
)

const (
	messageAttributeNameAll   = "All"
	messageAttributeNameTopic = "Topic"
)

// MessageProcessor reads and processes messages off of an AWS SQS queue.
type MessageProcessor struct {
	client                   *sqs.Client
	controller               types.Controller
	maxNumberOfMessages      int32
	maxWaitTimeSeconds       int32
	queueURL                 string
	visibilityTimeoutSeconds int32
}

// NewMessageProcessor creates a new instance of MessageProcessor.
func NewMessageProcessor(
	sqsClient *sqs.Client,
	sqsQueueURL string,
	sqsQueueMaxNumberOfMessages int32,
	sqsQueueMaxWaitTimeSeconds int32,
	sqsQueueVisibilityTimeoutSeconds int32,
	controller types.Controller,
) *MessageProcessor {
	return &MessageProcessor{
		client:                   sqsClient,
		controller:               controller,
		maxNumberOfMessages:      sqsQueueMaxNumberOfMessages,
		maxWaitTimeSeconds:       sqsQueueMaxWaitTimeSeconds,
		queueURL:                 sqsQueueURL,
		visibilityTimeoutSeconds: sqsQueueVisibilityTimeoutSeconds,
	}
}

// Run sits on a loop reading and processing messages off of the AWS SQS queue.
func (p *MessageProcessor) Run() {
	for {
		r, err := p.client.ReceiveMessage(
			context.TODO(),
			&sqs.ReceiveMessageInput{
				MaxNumberOfMessages: p.maxNumberOfMessages,
				MessageAttributeNames: []string{
					messageAttributeNameAll,
				},
				QueueUrl:          aws.String(p.queueURL),
				VisibilityTimeout: p.visibilityTimeoutSeconds,
				WaitTimeSeconds:   p.maxWaitTimeSeconds,
			},
		)

		if err != nil {
			log.Errorf("Failed to receive message: %v", err)
			continue
		}

		if len(r.Messages) <= 0 {
			continue
		}

		var wg sync.WaitGroup
		wg.Add(len(r.Messages))
		for _, message := range r.Messages {
			go func(message *sqsTypes.Message) {
				defer wg.Done()
				log.Tracef("Processing message with ID %q", *message.MessageId)

				var (
					body  string
					topic string
				)

				// Retrieve the message's body (if any).
				if message.Body == nil {
					body = ""
				} else {
					body = *message.Body
				}

				// Retrieve the message's topic (if any).
				if v, ok := message.MessageAttributes[messageAttributeNameTopic]; !ok || *v.StringValue == "" {
					topic = p.queueURL
				} else {
					topic = *v.StringValue
				}

				// Invoke the function(s) associated with the topic.
				log.Tracef("Invoking on topic %q passing message with ID %q", topic, *message.MessageId)
				p.controller.InvokeWithContext(
					buildMessageContext(message),
					topic,
					pointers.NewBytes([]byte(body)),
				)
			}(&message)
		}
		wg.Wait()
	}
}
