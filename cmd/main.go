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

package main

import (
	"context"
	"flag"
	"time"

	"github.com/Altitude-sports/connector-sdk/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	log "github.com/sirupsen/logrus"

	"github.com/Altitude-sports/openfaas-sqs-connector/internal/processors"
)

func main() {
	// Parse command-line flags
	logLevel := flag.String(
		"log-level",
		"info",
		"the log level to use",
	)
	maxNumberOfMessages := flag.Int(
		"max-number-of-messages",
		1,
		"the maximum number of messages to return from the aws sqs queue per iteration",
	)
	maxWaitTime := flag.Int(
		"max-wait-time",
		1,
		"the maximum amount of time (in seconds) to wait for messages to be returned from the aws sqs queue per iteration",
	)
	openfaasGatewayURL := flag.String(
		"openfaas-gateway-url",
		"",
		"the url at which the openfaas gateway can be reached",
	)
	queueURL := flag.String(
		"queue-url",
		"",
		"the URL of the AWS SQS queue to pop messages from",
	)
	queueName := flag.String(
		"queue-name",
		"",
		"the name of the AWS SQS queue to pop messages from",
	)
	awsAccountID := flag.String(
		"aws-account-id",
		"",
		"the AWS account ID that owns the AWS SQS queue to pop messages from",
	)
	region := flag.String(
		"region",
		"",
		"the AWS region where the SQS queue is located",
	)
	topicRefreshInterval := flag.Int(
		"topic-refresh-interval",
		15,
		"the interval (in seconds) at which to refresh the topic map",
	)
	visibilityTimeout := flag.Int(
		"visibility-timeout",
		30,
		"the amount of time (in seconds) during which received messages are unavailable to other consumers",
	)
	namespace := flag.String(
		"namespace",
		"",
		"the function namespace where it needs to be invoked",
	)
	flag.Parse()

	// Log at the requested level.
	if v, err := log.ParseLevel(*logLevel); err != nil {
		log.Fatalf("Failed to parse log level: %v", err)
	} else {
		log.SetLevel(v)
	}

	// Make sure that all required flags have been provided.
	if *queueURL == "" && *queueName == "" {
		log.Fatal("either --queue-url or --queue-name must be provided")
	}
	if *openfaasGatewayURL == "" {
		log.Fatal("--openfaas-gateway-url must be provided")
	}

	// Initialize the AWS SQS client.
	extraOpts := []func(*config.LoadOptions) error{}
	if *region != "" {
		extraOpts = append(extraOpts, config.WithRegion(*region))
	}

	awsConfig, err := config.LoadDefaultConfig(context.TODO(), extraOpts...)
	if err != nil {
		log.Fatalf("could not initialize the AWS SDK: %v\n", err)
	}

	sqsClient := sqs.NewFromConfig(awsConfig)

	// Retrieve the AWS SQS queue URL if not explicitly set
	if *queueURL == "" {
		opts := &sqs.GetQueueUrlInput{
			QueueName: aws.String(*queueName),
		}

		if *awsAccountID != "" {
			opts.QueueOwnerAWSAccountId = aws.String(*awsAccountID)
		}

		response, err := sqsClient.GetQueueUrl(context.TODO(), opts)
		if err != nil {
			log.Fatalf("could not retrieve the AWS SQS queue URL: %v\n", err)
		}

		queueURL = response.QueueUrl
	}

	// Initialize the controller.
	controller := types.NewController(
		types.GetCredentials(),
		&types.ControllerConfig{
			GatewayURL:        *openfaasGatewayURL,
			PrintResponse:     log.IsLevelEnabled(log.DebugLevel),
			PrintResponseBody: log.IsLevelEnabled(log.DebugLevel),
			PrintSync:         log.IsLevelEnabled(log.DebugLevel),
			RebuildInterval:   time.Duration(*topicRefreshInterval) * time.Second,
			Namespace:         *namespace,
		},
	)
	controller.BeginMapBuilder()

	// Initialize the response processor.
	controller.Subscribe(processors.NewResponseProcessor(sqsClient, *queueURL))

	// Initialize the message processor and start processing messages off the
	// AWS SQS queue.
	processors.NewMessageProcessor(
		sqsClient,
		*queueURL,
		int32(*maxNumberOfMessages),
		int32(*maxWaitTime),
		int32(*visibilityTimeout),
		controller,
	).Run()
}
