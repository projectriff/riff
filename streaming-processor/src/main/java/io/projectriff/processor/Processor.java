package io.projectriff.processor;

import com.github.bsideup.liiklus.protocol.AckRequest;
import com.github.bsideup.liiklus.protocol.Assignment;
import com.github.bsideup.liiklus.protocol.PublishRequest;
import com.github.bsideup.liiklus.protocol.ReactorLiiklusServiceGrpc;
import com.github.bsideup.liiklus.protocol.ReceiveReply;
import com.github.bsideup.liiklus.protocol.ReceiveRequest;
import com.github.bsideup.liiklus.protocol.SubscribeReply;
import com.github.bsideup.liiklus.protocol.SubscribeRequest;
import com.google.protobuf.ByteString;
import com.google.protobuf.Empty;
import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.Channel;
import io.grpc.netty.NettyChannelBuilder;
import io.projectriff.invoker.rpc.InputFrame;
import io.projectriff.invoker.rpc.InputSignal;
import io.projectriff.invoker.rpc.OutputFrame;
import io.projectriff.invoker.rpc.OutputSignal;
import io.projectriff.invoker.rpc.ReactorRiffGrpc;
import io.projectriff.invoker.rpc.StartFrame;
import io.projectriff.processor.serialization.Message;
import reactor.core.publisher.*;
import reactor.util.function.Tuple2;

import java.io.IOException;
import java.net.ConnectException;
import java.net.Socket;
import java.net.URI;
import java.net.URISyntaxException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.Duration;
import java.util.*;
import java.util.stream.Collectors;
import java.util.stream.IntStream;

/**
 * Main driver class for the streaming processor.
 *
 * <p>Continually pumps data from one or several input streams (see {@code riff-serialization.proto} for this so-called "at rest" format),
 * arranges messages in invocation windows and invokes the riff function over RPC by multiplexing messages from several
 * streams into one RPC channel (see {@code riff-rpc.proto} for the wire format).
 * On the way back, performs the opposite operations: de-muxes results and serializes them back to the corresponding
 * output streams.</p>
 *
 * @author Eric Bottard
 * @author Florent Biville
 */
public class Processor {

    /**
     * ENV VAR key holding the directory path where streams metadata can be found.
     */
    private static final String CNB_BINDINGS = "CNB_BINDINGS";

    /**
     * ENV VAR key holding the address of the function RPC, as a {@code host:port} string.
     */
    private static final String FUNCTION = "FUNCTION";

    /**
     * ENV VAR key holding the logical names for input parameter names, as a comma separated list of strings.
     */
    private static final String INPUT_NAMES = "INPUT_NAMES";

    /**
     * ENV VAR key holding the logical names for output result names, as a comma separated list of strings.
     */
    private static final String OUTPUT_NAMES = "OUTPUT_NAMES";

    /**
     * ENV VAR key holding the start offsets for each input stream, as a comma separated list of either "earliest" or "latest".
     */
    private static final String INPUT_START_OFFSETS = "INPUT_START_OFFSETS";

    /**
     * ENV VAR key holding the consumer group string this process should use.
     */
    private static final String GROUP = "GROUP";

    /**
     * The number of retries when testing http connection to the function.
     */
    private static final int NUM_RETRIES = 20;

    /**
     * Keeps track of a single gRPC stub per gateway address.
     */
    private final Map<String, ReactorLiiklusServiceGrpc.ReactorLiiklusServiceStub> liiklusInstancesPerAddress;

    /**
     * The ordered input streams for the function, in parsed form.
     */
    private final List<FullyQualifiedTopic> inputs;

    /**
     * The ordered output streams for the function, in parsed form.
     */
    private final List<FullyQualifiedTopic> outputs;

    /**
     * The ordered logical names for input parameters of the function.
     */
    private final List<String> inputNames;

    /**
     * For each input stream, whether to subscribe at earliest or latest offset.
     */
    private final List<String> startOffsets;

    /**
     * The ordered logical names for output results of the function.
     */
    private final List<String> outputNames;

    /**
     * The ordered list of expected content-types for function results.
     */
    private final List<String> outputContentTypes;

    /**
     * The consumer group string this process will use to identify itself when reading from the input streams.
     */
    private final String group;

    /**
     * This is used in a shutdown hook to force completion of the input signals Flux via takeUntilOther().
     */
    private UnicastProcessor killSignal = UnicastProcessor.create();

    /**
     * The RPC stub used to communicate with the function process.
     *
     * @see "riff-rpc.proto for the wire format and service definition"
     */
    private final ReactorRiffGrpc.ReactorRiffStub riffStub;

    public static void main(String[] args) throws Exception {

        long t0 = System.currentTimeMillis();

        checkEnvironmentVariables();

        Hooks.onOperatorDebug();

        String functionAddress = System.getenv(FUNCTION);

        List<String> inputNames = Arrays.asList(System.getenv(INPUT_NAMES).split(","));
        List<String> outputNames = Arrays.asList(System.getenv(OUTPUT_NAMES).split(","));

        List<FullyQualifiedTopic> inputAddressableTopics = resolveStreams(System.getenv(CNB_BINDINGS), "input", inputNames.size());
        List<FullyQualifiedTopic> outputAddressableTopics = resolveStreams(System.getenv(CNB_BINDINGS), "output", outputNames.size());
        List<String> outputContentTypes = resolveContentTypes(System.getenv(CNB_BINDINGS), outputNames.size());
        List<String> startOffsets = Arrays.asList(System.getenv(INPUT_START_OFFSETS).split(","));

        assertHttpConnectivity(functionAddress);
        Channel fnChannel = NettyChannelBuilder.forTarget(functionAddress)
                .usePlaintext()
                .build();

        Processor processor = new Processor(
                inputAddressableTopics,
                outputAddressableTopics,
                inputNames,
                startOffsets,
                outputNames,
                outputContentTypes,
                System.getenv(GROUP),
                ReactorRiffGrpc.newReactorStub(fnChannel));

        System.out.format("Connected to %s, after %d ms\n", functionAddress, System.currentTimeMillis() - t0);

        Runtime.getRuntime().addShutdownHook(new Thread() {
            @Override
            public void run() {
                processor.killSignal.sink().complete();
                try {
                    Thread.sleep(500);
                } catch (InterruptedException e) {
                }
            }
        });


        processor.run();

    }

    private static void checkEnvironmentVariables() {
        List<String> envVars = Arrays.asList(FUNCTION, GROUP, INPUT_NAMES, OUTPUT_NAMES, INPUT_START_OFFSETS, CNB_BINDINGS);
        if (envVars.stream()
                .anyMatch(v -> (System.getenv(v) == null || System.getenv(v).trim().length() == 0))) {
            System.err.format("Missing one of the following environment variables: %s%n", envVars);
            envVars.forEach(v -> System.err.format("  %s = %s%n", v, System.getenv(v)));
            System.exit(1);
        }
    }

    private static void assertHttpConnectivity(String functionAddress) throws URISyntaxException, IOException, InterruptedException {
        URI uri = new URI("http://" + functionAddress);
        for (int i = 1; i <= NUM_RETRIES; i++) {
            try (Socket s = new Socket(uri.getHost(), uri.getPort())) {
            } catch (ConnectException t) {
                if (i == NUM_RETRIES) {
                    throw t;
                }
                Thread.sleep(i * 100);
            }
        }
    }

    private Processor(List<FullyQualifiedTopic> inputs,
                      List<FullyQualifiedTopic> outputs,
                      List<String> inputNames,
                      List<String> startOffsets,
                      List<String> outputNames,
                      List<String> outputContentTypes,
                      String group,
                      ReactorRiffGrpc.ReactorRiffStub riffStub) {

        this.inputs = inputs;
        this.outputs = outputs;
        this.inputNames = inputNames;
        this.startOffsets = startOffsets;
        this.outputNames = outputNames;
        Set<FullyQualifiedTopic> allGateways = new HashSet<>(inputs);
        allGateways.addAll(outputs);

        this.liiklusInstancesPerAddress = indexByAddress(allGateways);
        this.outputContentTypes = outputContentTypes;
        this.riffStub = riffStub;
        this.group = group;
    }

    public static List<FullyQualifiedTopic> resolveStreams(String bindingsDir, String prefix, int count) {
        return IntStream.range(0, count)
                .mapToObj(i -> resolveStream(bindingsDir, String.format("%s_%03d", prefix, i)))
                .collect(Collectors.toList());
    }

    private static FullyQualifiedTopic resolveStream(String bindingsDir, String path) {
        Path root = Paths.get(bindingsDir).resolve(path).resolve("secret");
        try {
            String gateway = new String(Files.readAllBytes(root.resolve("gateway")), StandardCharsets.UTF_8);
            String topic = new String(Files.readAllBytes(root.resolve("topic")), StandardCharsets.UTF_8);
            return new FullyQualifiedTopic(gateway, topic);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    public void run() {
        Flux.fromIterable(inputs).zipWithIterable(startOffsets)
                .flatMap(inputTopic -> {
                    ReactorLiiklusServiceGrpc.ReactorLiiklusServiceStub inputLiiklus = liiklusInstancesPerAddress.get(inputTopic.getT1().getGatewayAddress());
                    return inputLiiklus.subscribe(subscribeRequestForInput(inputTopic))
                            .filter(SubscribeReply::hasAssignment)
                            .map(SubscribeReply::getAssignment)
                            .flatMap(
                                    assignment -> inputLiiklus
                                            .receive(receiveRequestForAssignment(assignment))
                                            .delayUntil(receiveReply -> ack(inputTopic.getT1(), inputLiiklus, receiveReply, assignment))
                            )
                            .map(receiveReply -> toRiffSignal(receiveReply, inputTopic.getT1()));
                })
                .takeUntilOther(killSignal)
                .transform(this::riffWindowing)
                .map(this::invoke)
                .concatMap(flux ->
                        flux.concatMap(m -> {
                            OutputFrame next = m.getData();
                            FullyQualifiedTopic output = outputs.get(next.getResultIndex());
                            ReactorLiiklusServiceGrpc.ReactorLiiklusServiceStub outputLiiklus = liiklusInstancesPerAddress.get(output.getGatewayAddress());
                            return outputLiiklus.publish(createPublishRequest(next, output.getTopic()));
                        })
                )
                .blockLast();
    }

    private Mono<Empty> ack(FullyQualifiedTopic topic, ReactorLiiklusServiceGrpc.ReactorLiiklusServiceStub stub, ReceiveReply receiveReply, Assignment assignment) {
        System.out.format("ACKing %s for group %s: offset=%d, part=%d%n", topic.getTopic(), this.group, receiveReply.getRecord().getOffset(), assignment.getPartition());
        return stub.ack(AckRequest.newBuilder()
                .setGroup(this.group)
                .setOffset(receiveReply.getRecord().getOffset())
                .setPartition(assignment.getPartition())
                .setTopic(topic.getTopic())
                .build());
    }

    private static Map<String, ReactorLiiklusServiceGrpc.ReactorLiiklusServiceStub> indexByAddress(
            Collection<FullyQualifiedTopic> fullyQualifiedTopics) {
        return fullyQualifiedTopics.stream()
                .map(FullyQualifiedTopic::getGatewayAddress)
                .distinct()
                .collect(Collectors.toMap(
                        address -> address,
                        address -> ReactorLiiklusServiceGrpc.newReactorStub(
                                NettyChannelBuilder.forTarget(address)
                                        .usePlaintext()
                                        .build())
                        )
                )
                ;
    }

    private Flux<OutputSignal> invoke(Flux<InputFrame> in) {
        InputSignal start = InputSignal.newBuilder()
                .setStart(StartFrame.newBuilder()
                        .addAllExpectedContentTypes(this.outputContentTypes)
                        .addAllInputNames(this.inputNames)
                        .addAllOutputNames(this.outputNames)
                        .build())
                .build();

        return riffStub.invoke(Flux.concat(
                Flux.just(start), //
                in.map(frame -> InputSignal.newBuilder().setData(frame).build())));
    }

    /**
     * This converts an RPC representation of an {@link OutputFrame} to an at-rest {@link Message}, and creates a publish request for it.
     */
    private PublishRequest createPublishRequest(OutputFrame next, String topic) {
        Message msg = Message.newBuilder()
                .setPayload(next.getPayload())
                .setContentType(next.getContentType())
                .putAllHeaders(next.getHeadersMap())
                .build();

        return PublishRequest.newBuilder()
                .setValue(msg.toByteString())
                .setTopic(topic)
                .build();
    }

    private static ReceiveRequest receiveRequestForAssignment(Assignment assignment) {
        return ReceiveRequest.newBuilder().setAssignment(assignment).build();
    }

    private <T> Flux<Flux<T>> riffWindowing(Flux<T> linear) {
        return linear.window(Duration.ofSeconds(60));
    }

    /**
     * This converts a liiklus received message (representing an at-rest riff {@link Message}) into an RPC {@link InputFrame}.
     */
    private InputFrame toRiffSignal(ReceiveReply receiveReply, FullyQualifiedTopic fullyQualifiedTopic) {
        int inputIndex = inputs.indexOf(fullyQualifiedTopic);
        if (inputIndex == -1) {
            throw new RuntimeException("Unknown topic: " + fullyQualifiedTopic);
        }
        ByteString bytes = receiveReply.getRecord().getValue();
        try {
            Message message = Message.parseFrom(bytes);
            return InputFrame.newBuilder()
                    .setPayload(message.getPayload())
                    .setContentType(message.getContentType())
                    .setArgIndex(inputIndex)
                    .build();
        } catch (InvalidProtocolBufferException e) {
            throw new RuntimeException(e);
        }

    }

    private SubscribeRequest subscribeRequestForInput(Tuple2<FullyQualifiedTopic, String> topicAddressAndOffset) {
        return SubscribeRequest.newBuilder()
                .setTopic(topicAddressAndOffset.getT1().getTopic())
                .setGroup(group)
                .setAutoOffsetReset(topicAddressAndOffset.getT2().equals("earliest") ? SubscribeRequest.AutoOffsetReset.EARLIEST : SubscribeRequest.AutoOffsetReset.LATEST)
                .build();
    }

    private static List<String> resolveContentTypes(String bindingsDir, int count) throws IOException {
        List<String> result = new ArrayList<>();
        for (int i = 0; i < count; i++) {
            Path path = Paths.get(bindingsDir).resolve(String.format("output_%03d", i)).resolve("metadata").resolve("contentType");
            result.add(new String(Files.readAllBytes(path), StandardCharsets.UTF_8));
        }
        return result;
    }
}