package io.sk8s.topic.gateway;

import java.util.HashMap;
import java.util.Iterator;
import java.util.Map;

import org.springframework.cloud.sleuth.Span;
import org.springframework.cloud.sleuth.SpanTextMap;
import org.springframework.cloud.sleuth.Tracer;
import org.springframework.cloud.sleuth.instrument.messaging.MessagingSpanTextMapInjector;
import org.springframework.messaging.Message;
import org.springframework.messaging.support.GenericMessage;
import org.springframework.messaging.support.MessageBuilder;
import org.springframework.messaging.support.MessageHeaderAccessor;
import org.springframework.messaging.support.NativeMessageHeaderAccessor;
import org.springframework.util.StringUtils;

public class MessageTraceHandler {

	private Tracer tracer;

	private MessagingSpanTextMapInjector injector;

	public MessageTraceHandler(Tracer tracer, MessagingSpanTextMapInjector injector) {
		this.tracer = tracer;
		this.injector = injector;
	}

	public Message<?> traceBeforeSend(Message<?> message, String topicName) {

		MessageBuilder<?> messageBuilder = MessageBuilder.fromMessage(message);

		Span parentSpan = tracer.getCurrentSpan();
		Span span = tracer.createSpan(topicName, parentSpan);
		span.logEvent(Span.CLIENT_SEND);
		messageBuilder.setHeader("messageSent", true);

		injector.inject(span, new MessagingTextMap(messageBuilder));
		MessageHeaderAccessor headers = MessageHeaderAccessor.getMutableAccessor(message);
		headers.copyHeaders(messageBuilder.build().getHeaders());
		return new GenericMessage<>(message.getPayload(), headers.getMessageHeaders());
	}

	public void traceAfterSend(Message<?> message) {
		Span currentSpan = tracer.getCurrentSpan();
		currentSpan.logEvent(Span.SERVER_SEND);
		tracer.close(currentSpan);
	}

	class MessagingTextMap implements SpanTextMap {

		private final MessageBuilder delegate;

		public MessagingTextMap(MessageBuilder delegate) {
			this.delegate = delegate;
		}

		@Override
		public Iterator<Map.Entry<String, String>> iterator() {
			Map<String, String> map = new HashMap<>();
			for (Map.Entry<String, Object> entry : this.delegate.build().getHeaders()
					.entrySet()) {
				map.put(entry.getKey(), String.valueOf(entry.getValue()));
			}
			return map.entrySet().iterator();
		}

		@SuppressWarnings("unchecked")
		public void put(String key, String value) {
			if (!StringUtils.hasText(value)) {
				return;
			}
			Message<?> initialMessage = this.delegate.build();
			MessageHeaderAccessor accessor = MessageHeaderAccessor
					.getMutableAccessor(initialMessage);
			accessor.setHeader(key, value);
			if (accessor instanceof NativeMessageHeaderAccessor) {
				NativeMessageHeaderAccessor nativeAccessor = (NativeMessageHeaderAccessor) accessor;
				nativeAccessor.setNativeHeader(key, value);
			}
			this.delegate.copyHeaders(accessor.toMessageHeaders());
		}
	}
}
