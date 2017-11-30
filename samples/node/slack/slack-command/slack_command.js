module.exports = function(input) {
	/*
	 * This will echo plain text or the contents of the text property, which
	 * is what slack command POSTs.
	 */
	 return "received:" + (typeof(input) == "string" ? input : input.text);
};
