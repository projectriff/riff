const express = require('express');
const bodyParser = require('body-parser');

const PORT = 8080;

const app = express();

app.use('/', bodyParser.text());

// Download function
var wget = require('node-wget');
wget(process.env.FUNCTION_URI);
var i = process.env.FUNCTION_URI.lastIndexOf("/");
var fnFileName = process.env.FUNCTION_URI.substr(i);

app.post('/', function (req, res) {
    var fn = require(fnFileName);
    var resultx = fn(req.body);
    console.log("Result " + resultx);
    res.type("text/plain");
    res.status(200).send(resultx);
});

app.listen(PORT);
console.log('Running on http://localhost:' + PORT);
