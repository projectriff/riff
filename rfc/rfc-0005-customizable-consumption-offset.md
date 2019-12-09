# RFC-0005: Optional starting offset for processor inputs

**Authors:** Eric Bottard

**Status:**

**Pull Request URL:** https://github.com/projectriff/riff/pull/1367

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** rfc-0003: developer utils, which also support viewing stream data from the earliest offset.


## Problem
Sometimes users want to replay data through functions, maybe because the function logic has changed, or maybe because downstream systems have lost state and that state needs to be re-created. The current Processor concept in riff doesn't support such a mechanism though, and always consumes data from the (current) end of streams.

### Anti-Goals

## Solution
With a top to bottom view of the user experience, the proposed solution is to
- introduce an offset option on each input binding when a processor is created
- persist that option in the Processor CRD instance on the cluster
- have the Processor reconciler forward that option to the streaming-processor, so that it changes its SubscriptionRequests accordingly

### User Impact
This RFC proposes to change the `Inputs` field of the `Processor` CRD from
```go
Inputs []StreamBinding `json:"inputs"`
```
to
```go
Inputs []InputStreamBinding `json:"inputs"`
...
type InputStreamBinding struct {
    Stream string `json:"stream"`
    Alias string `json:"alias,omitempty"`

	// Where to start consuming this stream the first time a processor runs.
    StartOffset string `json:"startOffset"`
}
```

with `StartOffset` supporting only two special, textual values for now: `earliest` and `latest` (default). Future support for parsing to a non negative number after failing to recognize those two values is possible.

This RFC proposes to change the `riff streaming processor create` command from
```
riff streaming processor create <name> ... --input [alias:]stream ...
```
to
```
riff streaming processor create <name> ... --input [alias:]stream[:offset] ...
```
where `[..]` denotes optional values and `[:offset]` could take values `earliest` or `latest`. Those two logical values are the only ones currently supported by liiklus (and the high level kafka consumer as far as the author can remember). Nevertheless, the encoding of those two values could be done via negative numerical values, in preparation of future support of an actual numerical index.


The actual behavior of this proposed change has to be well understood: the offset parameter only makes sense for a _new consumer group_, _i.e._ a group which has never been used and for which liiklus doesn't currently have positions. As soon as the group starts to be used and its positions stored, the `earliest` value will have no effect. Indeed, this is desired behavior otherwise everytime a processor scales down and up again, it would replay everything over and over again. The corollary from this is that to start consuming from the beginning *again*, a processor would need to be *re-created*, with a different name (because the consumer group is based on the name of the processor).

NOTE: riff currently uses in-memory positions storage for liiklus, until support for Kafka storage is supported. So technically, restarting the liiklus process of a given provider would allow starting from a blank slate.

### Backwards Compatibility and Upgrade Path
The proposed change should be totally backwards compatible with existing riff with
- the cli option being optional
- its default value being `latest`, which is the current effective mode used.

## FAQ
*Answers to common questions that may arise and those that you’ve commonly been asked after requesting comments for this proposal.*
