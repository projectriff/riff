package functions;

import java.util.function.Function;

public class Upper implements Function<String, String> {

	public String apply(String name) {
		return name.toUpperCase();
	}
}
