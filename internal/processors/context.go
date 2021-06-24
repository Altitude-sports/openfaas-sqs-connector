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

	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	log "github.com/sirupsen/logrus"
)

type contextKey int

const (
	loggerContextKey contextKey = iota
	messageIDContextKey
	messageReceiptHandleContextKey
)

func buildMessageContext(m *sqsTypes.Message) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, loggerContextKey, log.WithField("message_id", *m.MessageId))
	ctx = context.WithValue(ctx, messageIDContextKey, *m.MessageId)
	ctx = context.WithValue(ctx, messageReceiptHandleContextKey, *m.ReceiptHandle)
	return ctx
}

func unpackMessageContext(ctx context.Context) (
	logEntry *log.Entry,
	messageID string,
	messageReceiptHandle string,
) {
	logEntry = ctx.Value(loggerContextKey).(*log.Entry)
	messageID = ctx.Value(messageIDContextKey).(string)
	messageReceiptHandle = ctx.Value(messageReceiptHandleContextKey).(string)
	return
}
