/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package io.sk8s.sidecar;

import org.springframework.retry.backoff.ExponentialBackOffPolicy;
import org.springframework.retry.policy.SimpleRetryPolicy;
import org.springframework.retry.support.RetryTemplate;
import org.springframework.web.client.RestTemplate;

import javax.annotation.PostConstruct;

/**
 * @author Mark Fisher
 */
public class HttpDispatcher implements Dispatcher {

    private static final String INVOKER_URL = "http://localhost:8080/";

    private final RestTemplate restTemplate = new RestTemplate();
    private final RetryTemplate retryTemplate = new RetryTemplate();
    private final ExponentialBackOffPolicy backOffPolicy = new ExponentialBackOffPolicy();
    private final SimpleRetryPolicy retryPolicy = new SimpleRetryPolicy();

    @Override
    public String dispatch(String input) {
        return this.retryTemplate.execute(retryContext -> restTemplate.postForObject(INVOKER_URL, input, String.class));
    }

    @PostConstruct
    public void initIt() {
        this.backOffPolicy.setInitialInterval(1);
        this.backOffPolicy.setMaxInterval(5);
        this.retryPolicy.setMaxAttempts(3);
        this.retryTemplate.setBackOffPolicy(backOffPolicy);
        this.retryTemplate.setRetryPolicy(retryPolicy);
    }

}
