# RFC-0013: Invoker TCK

**Authors:** Eric Bottard

**Status:** Accepted

**Pull Request URL:** https://github.com/projectriff/riff/pull/1378

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** 


## Problem
A specification of how function invokers should produce a runnable function exists (available at https://github.com/projectriff/invoker-specification). While it does its best to be prescriptive about what an invoker should and should not do, there are subtle caveats about the expected behavior, in particular in the _streaming_ case. Having a compatibility kit that implementors can run to ensure compliance would help detecting problems upstream.

## Solution
This RFC proposes to introduce a TCK that would take the form of at least a runnable tool (maybe also available as a go library) that focuses on testing the conformance of packaged functions with the _interaction (networking) protocols_ prescribed by the invoker specification.
As such, it would require that users of the TCK produce a set of Docker images implementing a handful of trivial functions (such as a "square" function, see below). 
How these images got constructed is deemed out of scope of the TCK, and while riff and the specification propose using buildpacks to construct such images, the TCK focuses on the resulting behavior of the container, which is assumed to result from the usage of the _invoker layer_.
The use of containers as the unit of testing has several advantages:
- the overall environment needed for the invoker to function properly does not have to be setup as part of the templated run of a test. Invokers implementors are in full control to provide a runnable container, and there is no software conflict with the environment used to drive the TCK run.
- using a container guarantees isolation of test runs, in particular with regard to data left behind.

The TCK would roughly take the following form:
A series of tests run that cover the actual requirements of the specification. In particular, the words "MUST" throughout the spec should correspond to actual testcases. To run such tests, the invoker implementor is asked to provide a series of containers that implement simple functions. For example, to test the http request/reply interaction and correct support for marshalling/unmarshalling of `application/json`, the TCK may leverage a container that assumes input as a number and returns its square. The TCK runner would run the container image configured for "square", POST a payload on `http://<host>:8080/` and verify synchronous return of a squared number, with `Content-Type` compatible with the requested `Accept` header.
The TCK will devide its test suites into several groups, corresponding to several layers of conformance. For example, support for http request/reply interaction and gRPC streaming interaction would be tested separately and reported as such. Similarly, occurences of SHOULD or MAY in the spec will be covered by tests that report conformance but don't fail the run if the tests don't pass.

### User Impact
The term "user" here means implementors of a function invoker. The requirements for such users to be able to use the TCK would likely be:
- ability to run a compiled go command line tool,
- availability of a local Docker daemon.

### Backwards Compatibility and Upgrade Path
The RFC concerns implementors of function invokers and has no impact on end users.