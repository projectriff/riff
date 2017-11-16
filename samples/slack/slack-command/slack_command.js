module.exports = function(input) {
	var kvPairs = input.split("&").map(pair => pair.split("="));
	textPair = kvPairs.find(pair => pair[0] == "text");
	if (textPair && textPair.length == 2) {
		return textPair[1];
	}
	else {
		return "BARF: Here's the input you gave me: " + input;
	}
};
