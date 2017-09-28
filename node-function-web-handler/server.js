const express = require('express');
const bodyParser = require('body-parser');

const PORT = 8080;

const app = express();

app.use('/init', bodyParser.json());
app.use('/invoke', bodyParser.text());

var fn;

app.post('/init', function (req, res) {
    fn = require(req.body.uri);
    console.log("Loaded function from " + req.body.uri);
    console.log("Function takes " + fn.length + " arguments");
    res.type("text/plain");
    res.status(200).send("OK");
});

app.post('/invoke', function (req, res) {
    if (fn === undefined) {
        res.status(500).send("Function not yet initialized");
        return;
    }
	var resultx = fn(req.body);
    console.log("Result " + resultx);
    res.type("text/plain");
    res.status(200).send(resultx);
});

app.listen(PORT);
console.log('Running on http://localhost:' + PORT);
