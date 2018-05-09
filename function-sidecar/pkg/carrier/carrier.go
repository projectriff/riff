/*
 * Copyright 2018 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package carrier

import (
	"log"

	dispatch "github.com/projectriff/riff/function-sidecar/pkg/dispatcher"
	"github.com/projectriff/riff/message-transport/pkg/transport"
)

func Run(consumer transport.Consumer, producer transport.Producer, dispatcher dispatch.Dispatcher, replyTopic string) {

	go func() {
		for {
			// Incoming message
			msg, _, err := consumer.Receive()
			if err == nil {
				log.Printf(">>> %s\n", msg)
				dispatcher.Input() <- msg
			} else {
				// Transport closed
				log.Print("Exiting transport Consumer loop")
				return
			}
		}
	}()

	go func() {
		for {
			// Result message
			resultMsg, open := <-dispatcher.Output() // Make sure to drain channel even if output==""
			if open {
				if replyTopic != "" {
					log.Printf("<<< %s\n", resultMsg)
					err := producer.Send(replyTopic, resultMsg)
					if err != nil {
						log.Printf("Error sending reply message: %v", err)
						continue
					}
				} else {
					log.Printf("=== Not sending function return value as reply topic was not provided. Raw result = %s\n", resultMsg)
				}
			} else {
				log.Print("Exiting transport Producer loop")
				return
			}
		}
	}()

}
