/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *  
 *        http://www.apache.org/licenses/LICENSE-2.0
 *  
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package jsonpath

import (
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	"os"

)

var parser_np *Parser
var parser_lb *Parser

func setup() {
	dat, err := ioutil.ReadFile("gateway_np.json")
	if err != nil {
		panic(err)
	}
	parser_np = NewParser(dat)

	dat, err = ioutil.ReadFile("gateway_lb.json")
	if err != nil {
		panic(err)
	}
	parser_lb = NewParser(dat)
}

func shutdown() {

}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}
func TestParserWithLoadBalancer(t *testing.T) {
	as := assert.New(t)

	portType := parser_lb.Value(`$.items[0].spec.type+`)

	as.Equal("LoadBalancer", portType)

	ip := parser_lb.Value(`$.items[0].status.loadBalancer.ingress[0].ip+`)

	as.Equal("35.196.105.42", ip)

	nodePort := parser_lb.Value(`$.items[0].spec.ports[*]?(@.name == "http").nodePort+`)
	as.Equal("32627", nodePort)
}

func TestNodePortQuery(t *testing.T) {

	as := assert.New(t)

	portType := parser_np.Value(`$.items[0].spec.type+`)

	as.Equal("NodePort", portType)

	nodePort := parser_np.Value(`$.items[0].spec.ports[*]?(@.name == "http").nodePort+`)
	as.Equal("31861", nodePort)

	ip := parser_np.Value(`$.items[0].status.loadBalancer.ingress[0].ip+`)

	as.Equal("", ip)

}