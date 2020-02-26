# nop-provisioner

The Nop Provisioner component is responsible for returning an address to
a gateway when asked to do so by the [riff system](https://github.com/projectriff/system).

For a given riff `stream` "foo" existing in namespace "my-ns",
a PUT request will be made to this component at `/my-ns/foo`.
Unlike other provisioners that create resources, this merely
returns its coordinates.

Upon success,
it will return its [liiklus](https://github.com/bsideup/liiklus)
coordinates in the following json form:
```json
{
  "gateway": "<host>:<port>",
  "topic": "<assigned-topic-name>"
}
```

## Configuration
The provisioner should run with the following environment variables
configured:

* `GATEWAY`: the address of a liiklus gRPC endpoint. Will be used as part
of the returned coordinates (see above).

