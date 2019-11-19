# RFC-001: CLI support for publishing and subscribing to streams

**To be Reviewed By:** 11/19/2019

**Authors:** Swapnil Bawaskar

**Status:** **Active** | Accepted | Dropped | Superseded

**Superseded by:** N/A

**Related:** N/A


## Problem
When the streaming runtime is installed, riff users can create streams (backed by a stream provider e.g. kafka). Users can write functions that read from and write to these streams,
but there is no easy mechanism to get events into the stream and look into the contents of the stream which are helpful while developing/demoing.

### Anti-Goals
We will only address this problem for development/demos, not production, so topics like auth/authz are out of scope for this document.

## Solution
The idea is to provide users with only one cli (riff cli) that will enable users to send/receive messages to/from a stream (commands below). We will have to ask the users to port-forward the gateway service if they are running outside the cluster.

#### User Impact
We should introduce the following commands to the riff cli:
- send a message to the stream:

    `riff streaming stream publish <stream-name> --payload <payload-as-string> --headers <headers-as-string>`  
    content type has already been specified while creating the stream.

- subscribe for messages from a stream:

    `riff streaming stream subscribe <stream-name> --offset <long-offset>`  
    this will display all the message from the given offset and then block to display further messages, essentially preventing user from entering subsequent commands.
    The messages will be displayed as json with the following fields:
    ``` 
    {
        "payload": "the users payload",
        "content-type": "the content type of the message",
        "headers": {"user provided header": "while publishing"}
    }
    ```

### Interim solution:
We have a go library at projectriff/stream-client-go which has methods to `Publish` and `Subscribe` to/from streams. We will create two simple binaries one for publishing and one for subscribing within this repo.

Steps:
1. Move the existing client.go and client_test.go to `pkg` folder. 
1. create `cmd/publish/main.go` for the publish binary
1. create `cmd/subscribe/main.go` for the subscribe binary.
1. setup CI to publish the binaries.

#### User Impact
The users will download the publish and subscribe binaries and run them by passing positional parameters:  
- `$ streampublish <gateway-address> <topic-name> <payload-as-string> <key> <headers> <content-type>`
- `$ streamsubscribe <gateway-address> <topic-name> <group> <long-offset>`

### Backwards Compatibility and Upgrade Path

## FAQ
*Answers to questions you’ve commonly been asked after requesting comments for this proposal.*
