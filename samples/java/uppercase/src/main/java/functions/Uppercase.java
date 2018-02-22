package functions;

import java.util.function.Function;

public class Uppercase implements Function<String, String> {

	@Override
	public String apply(String s) {
		return s.toUpperCase();
	}
}
