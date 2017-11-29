/*
 * Copyright 2017 the original author or authors.
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

package pipes

import (
	"bufio"
	"errors"
	"github.com/projectriff/function-sidecar/pkg/dispatcher"
	"log"
	"os"
	"strings"
	"syscall"
)

const (
	INPUT_PIPE  = "/pipes/input"
	OUTPUT_PIPE = "/pipes/output"
)

type pipesDispatcher struct {
	input  chan<- dispatcher.Message
	output <-chan dispatcher.Message
}

func (this pipesDispatcher) Input() chan<- dispatcher.Message {
	return this.input
}

func (this pipesDispatcher) Output() <-chan dispatcher.Message {
	return this.output
}

func NewPipesDispatcher() (dispatcher.Dispatcher, error) {
	log.Println("Creating new pipes Dispatcher")
	err := syscall.Mkfifo(INPUT_PIPE, 0666)
	if err != nil {
		log.Printf("error creating input pipe: %v", err)
		err = os.Remove(INPUT_PIPE)
		if err != nil {
			log.Printf("error removing input pipe: %v", err)
			return nil, err
		}
		err = syscall.Mkfifo(INPUT_PIPE, 0666)
		if err != nil {
			return nil, err
		}
	} else {
		log.Printf("Created %v\n", INPUT_PIPE)
	}
	err = syscall.Mkfifo(OUTPUT_PIPE, 0666)
	if err != nil {
		err = os.Remove(OUTPUT_PIPE)
		if err != nil {
			return nil, err
		}
		err = syscall.Mkfifo(OUTPUT_PIPE, 0666)
		if err != nil {
			return nil, err
		}
	} else {
		log.Printf("Created %v\n", OUTPUT_PIPE)
	}

	outfile, err := os.OpenFile(INPUT_PIPE, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(outfile)
	infile, err := os.OpenFile(OUTPUT_PIPE, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(infile)

	i := make(chan dispatcher.Message, 100) // TODO buffered
	o := make(chan dispatcher.Message, 100)

	go func() {
		for {
			select {
			case in, open := <-i:
				if open {
					err := writeMessage(writer, in)
					if err != nil {
						log.Printf("Error writing to %v: %v", INPUT_PIPE, err)
						break
					}
				} else {
					close(o)
					log.Print("Shutting down pipes dispatcher")
					return
				}
			}
		}
	}()

	go func() {
		for {
			msg, err := readMessage(reader)
			if err != nil {
				log.Printf("error reading message: %v", err)
				continue
			}
			if msg != nil {
				o <- *msg
			}
		}
	}()

	return pipesDispatcher{input: i, output: o}, nil
}

func readMessage(reader *bufio.Reader) (*dispatcher.Message, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	} else if line != "#headers\n" {
		return nil, errors.New("Unexpected first line of message: " + line)
	}

	message := dispatcher.Message{}
	for line, err = reader.ReadString('\n'); line != "#payload\n"; line, err = reader.ReadString('\n') {
		if err != nil {
			return nil, err
		}
		line = line[0: len(line)-1] // drop CR
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			return nil, errors.New("Malformed header line: " + line)
		}
		if message.Headers == nil {
			message.Headers = make(map[string]interface{})
		}
		message.Headers[kv[0]] = kv[1]
	}

	payload := make([]byte, 0, 256)
	for line, err = reader.ReadString('\n'); line != "#end\n"; line, err = reader.ReadString('\n') {
		if err != nil {
			return nil, err
		}
		payload = append(payload, []byte(line)...)
	}
	// Strip last CR
	payload = payload[0: len(payload)-1]

	message.Payload = payload
	return &message, nil
}

func writeMessage(writer *bufio.Writer, in dispatcher.Message) error {
	_, err := writer.WriteString("#headers\n")
	if err != nil {
		return err
	}
	if in.Headers != nil {
		for k, v := range in.Headers {
			_, err = writer.WriteString(k + "=" + v.(string) + "\n")
			if err != nil {
				return err
			}
		}
	}
	_, err = writer.WriteString("#payload\n")
	if err != nil {
		return err
	}
	_, err = writer.WriteString(string(in.Payload.([]byte)) + "\n")
	if err != nil {
		return err
	}
	_, err = writer.WriteString("#end\n")
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
