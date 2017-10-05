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

import java.io.File;
import java.io.FileReader;
import java.io.FileWriter;
import java.io.LineNumberReader;

/**
 * @author Mark Fisher
 */
public class StdioDispatcher implements Dispatcher {

	private static final String INPUT_PIPE = "/pipes/input";

	private static final String OUTPUT_PIPE = "/pipes/output";

	private final FileWriter writer;

	private final LineNumberReader reader;

	public StdioDispatcher() {
		try {
			Process p1 = Runtime.getRuntime().exec(new String[]{"mkfifo", INPUT_PIPE});
			Process p2 = Runtime.getRuntime().exec(new String[]{"mkfifo", OUTPUT_PIPE});
			p1.waitFor();
			p2.waitFor();
			this.writer = new FileWriter(new File(INPUT_PIPE));
			this.reader = new LineNumberReader(new FileReader(new File(OUTPUT_PIPE)));
		}
		catch (Exception e) {
			throw new IllegalStateException("failed to create named pipes", e);
		}
	}

	@Override
	public String dispatch(String input) throws Exception {
		this.writer.write(input + "\n");
		this.writer.flush();
		return this.reader.readLine();
	}
}
