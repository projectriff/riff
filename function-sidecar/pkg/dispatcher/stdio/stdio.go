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
	"fmt"
	"github.com/sk8sio/function-sidecar/pkg/dispatcher"
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

func (this stdioDispatcher) Dispatch(in interface{}) (interface{}, error) {
	_, err := this.writer.WriteString(in.(string) + "\n")
	if err != nil {
		log.Printf("Error writing to %v: %v", OUTPUT_PIPE, err)
		return nil, err
	}
	err = this.writer.Flush()
	//fmt.Println("Wrote " + in.(string))
	if err != nil {
		log.Printf("Error flushing %v", err)
		return nil, err
	}
	line, err := this.reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading from %v: %v", INPUT_PIPE, err)
		return nil, err
	}
	//fmt.Println("Read " + line)
	return line[0 : len(line)-1], nil
}

func NewStdioDispatcher() dispatcher.Dispatcher {
	fmt.Println("Creating new stdio Dispatcher")
	err := syscall.Mkfifo(INPUT_PIPE, 0666)
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("Created %v\n", INPUT_PIPE)
	}
	err = syscall.Mkfifo(OUTPUT_PIPE, 0666)
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("Created %v\n", OUTPUT_PIPE)
	}

	result := stdioDispatcher{}

	outfile, err := os.OpenFile(INPUT_PIPE, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}
	result.writer = bufio.NewWriter(outfile)
	infile, err := os.OpenFile(OUTPUT_PIPE, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}
	result.reader = bufio.NewReader(infile)

	return result
}
