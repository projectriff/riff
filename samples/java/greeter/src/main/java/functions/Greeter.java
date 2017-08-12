package functions;

import java.util.function.Function;

public class Greeter implements Function<String, String> {

	public String apply(String name) {
		return "Hello " + name;
	}
}
