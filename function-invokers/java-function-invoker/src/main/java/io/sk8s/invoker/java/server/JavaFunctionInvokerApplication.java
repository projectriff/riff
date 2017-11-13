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

package io.sk8s.invoker.java.server;

import java.io.File;
import java.net.URL;
import java.net.URLClassLoader;
import java.util.ArrayList;
import java.util.List;
import java.util.jar.JarFile;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.util.StringUtils;

/**
 * @author Mark Fisher
 */
@SpringBootApplication
public class JavaFunctionInvokerApplication {

	private static Log logger = LogFactory.getLog(JavaFunctionInvokerApplication.class);

	public static void main(String[] args) {
		JavaFunctionInvokerApplication application = new JavaFunctionInvokerApplication();
		if (application.isolated(args)) {
			application.runner().run(args);
		}
		else {
			SpringApplication.run(JavaFunctionInvokerApplication.class, args);
		}
	}

	ApplicationRunner runner() {
		ApplicationRunner runner = new ApplicationRunner(createClassLoader(),
				JavaFunctionInvokerApplication.class.getName());
		return runner;
	}

	private boolean isolated(String[] args) {
		for (String arg : args) {
			if (arg.equals("--function.runner.isolated=true")) {
				return true;
			}
		}
		return false;
	}

	private URLClassLoader createClassLoader() {
		ClassLoader base = getClass().getClassLoader();
		if (!(base instanceof URLClassLoader)) {
			throw new IllegalStateException("Need a URL class loader, found: " + base);
		}
		@SuppressWarnings("resource")
		URLClassLoader urlClassLoader = (URLClassLoader) base;
		URL[] urls = urlClassLoader.getURLs();
		if (urls.length == 1) {
			URL[] classpath = extractClasspath(urls[0]);
			if (classpath != null) {
				urls = classpath;
			}
		}
		List<URL> child = new ArrayList<>();
		List<URL> parent = new ArrayList<>();
		for (URL url : urls) {
			child.add(url);
		}
		for (URL url : urls) {
			if (isRoot(StringUtils.getFilename(clean(url.toString())))) {
				parent.add(url);
				child.remove(url);
			}
		}
		logger.info("Parent: " + parent);
		logger.info("Child: " + child);
		if (!parent.isEmpty()) {
			base = new URLClassLoader(parent.toArray(new URL[0]), base.getParent());
		}
		return new URLClassLoader(child.toArray(new URL[0]), base);
	}

	private boolean isRoot(String file) {
		return file.startsWith("reactor-core") || file.startsWith("reactive-streams");
	}

	private String clean(String jar) {
		// This works with fat jars like Spring Boot where the path elements look like
		// jar:file:...something.jar!/.
		return jar.endsWith("!/") ? jar.substring(0, jar.length() - 2) : jar;
	}

	private URL[] extractClasspath(URL url) {
		// This works for a jar indirection like in surefire and IntelliJ
		if (url.toString().endsWith(".jar")) {
			JarFile jar;
			try {
				jar = new JarFile(new File(url.toURI()));
				String path = jar.getManifest().getMainAttributes()
						.getValue("Class-Path");
				if (path != null) {
					List<URL> result = new ArrayList<>();
					for (String element : path.split(" ")) {
						result.add(new URL(element));
					}
					return result.toArray(new URL[0]);
				}
			}
			catch (Exception e) {
			}
		}
		return null;
	}

}