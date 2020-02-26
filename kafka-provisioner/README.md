# kafka-provisioner

The Kafka Provisioner component is responsible for creating
topics in a Kafka cluster when asked to do so by the 
[riff system](https://github.com/projectriff/system).

For a given riff `stream` "foo" existing in namespace "my-ns",
a PUT request will be made to this component at `/my-ns/foo`.
It will react by creating a topic named `my-ns_foo`.
Note that the underscore is an allowed character in Kafka topic names
but is disallowed is kubernetes resource names.
This avoids collisions between `ns=foo-bar:stream=quizz` and 
`ns=foo:stream=bar-quizz` 

Upon successful creation (or lookup of pre-existing) of a topic,
it will return its [liiklus](https://github.com/bsideup/liiklus)
coordinates in the following json form:
 ```json
{
  "gateway": "<host>:<port>",
  "topic": "<created-topic-name>"
}
```

## Configuration
The provisioner should run with the following environment variables
configured:
* `BROKER`: the address of a Kafka broker to connect to, in the form `host:port`
* `GATEWAY`: the address of a liiklus gRPC endpoint. Will be used as part
of the returned coordinates (see above).

