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