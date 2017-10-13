const express = require('express');
const bodyParser = require('body-parser');

const PORT = 8080;

const app = express();

app.use('/', bodyParser.text());

var fn = require(process.env.FUNCTION_URI);

app.post('/', function (req, res) {
    var resultx = fn(req.body);
    console.log("Result " + resultx);
    res.type("text/plain");
    res.status(200).send(resultx);
});

app.listen(PORT);
console.log('Running on http://localhost:' + PORT);
