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

package stdio

import (
	"bufio"
	"github.com/projectriff/function-sidecar/pkg/dispatcher"
	"github.com/projectriff/message-transport/pkg/message"
	"log"
	"os"
	"syscall"
)

const (
	INPUT_PIPE  = "/pipes/input"
	OUTPUT_PIPE = "/pipes/output"
)

type stdioDispatcher struct {
	reader *bufio.Reader
	writer *bufio.Writer
}

func (this stdioDispatcher) Dispatch(in message.Message) (message.Message, error) {
	_, err := this.writer.WriteString(string(in.Payload()) + "\n")
	if err != nil {
		log.Printf("Error writing to %v: %v", OUTPUT_PIPE, err)
		return nil, err
	}
	err = this.writer.Flush()
	if err != nil {
		log.Printf("Error flushing %v", err)
		return nil, err
	}
	line, err := this.reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading from %v: %v", INPUT_PIPE, err)
		return nil, err
	}
	return message.NewMessage([]byte(line[0 : len(line)-1]), nil), nil
}

func (this stdioDispatcher) Close() error {
	err1 := os.Remove(INPUT_PIPE)
	err2 := os.Remove(OUTPUT_PIPE)
	log.Printf("err1 = %v"  , err1)
	log.Printf("err2 = %v"  , err2)
	if err1 == nil {
		return err2
	} else {
		return err1
	}
}

func NewStdioDispatcher() (dispatcher.SynchDispatcher, error) {
	log.Println("Creating new stdio Dispatcher")
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

	result := stdioDispatcher{}

	outfile, err := os.OpenFile(INPUT_PIPE, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}
	result.writer = bufio.NewWriter(outfile)
	infile, err := os.OpenFile(OUTPUT_PIPE, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}
	result.reader = bufio.NewReader(infile)

	return result, nil
}
