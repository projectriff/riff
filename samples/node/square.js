module.exports = function(input) {
	var n = parseFloat(input);
	if (isNaN(n) ) {
		return input + " is not a number";
	} else {
        return "The square of " + input + " is " + input*input;
	}
};
