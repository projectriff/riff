# pulsar-provisioner

The Pulsar Provisioner component is responsible for assigning
topics in a [Pulsar](https://pulsar.apache.org/) cluster when asked to do so by the 
[riff system](https://github.com/projectriff/system).

For a given riff `stream` "foo" existing in namespace "my-ns",
a PUT request will be made to this component at `/my-ns/foo`.
It will react by assigning a persistent topic named `persistent://$TENANT/$NAMESPACE/my-ns_foo`
in Pulsar (see below for configuration of `$TENANT` and `$NAMESPACE`, Pulsar namespaces are distinct from Kubernetes namespaces).
Note that because topics are created on demand in Pulsar, and because
this provisioner does not yet support advanced topic configuration options,
it does not _actually_ create the topic. It merely returns its coordinates.

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

* `TENANT`: the Pulsar [tenant](https://pulsar.apache.org/docs/en/concepts-multi-tenancy/#tenants) to provision topics in
* `NAMESPACE`: the Pulsar [namespace](https://pulsar.apache.org/docs/en/concepts-multi-tenancy/#namespaces) to provision topics in
* `BROKER`: the address of a Pulsar cluster to connect to, in the form `pulsar://host:port` (Currently unused)
* `GATEWAY`: the address of a liiklus gRPC endpoint. Will be used as part
of the returned coordinates (see above).

