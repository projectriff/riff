Testing locally:

```
$ mvn install dockerfile:build
$ docker run -ti -p 8080:8080 -v `pwd`/target/test-classes:/classes sk8s/java-function-invoker:0.0.1-SNAPSHOT --function.uri=file:classes?io.sk8s.invoker.java.function.Doubler
```

Then

```
$ curl -v localhost:8080 -H "Content-Type: text/plain" -d 5
10
```

For a function of `Flux` it only works if you explicitly start the app with `--function.runner.isolated=true`, i.e:

```
$ docker run -ti -p 8080:8080 -v `pwd`/target/test-classes:/classes sk8s/java-function-invoker:0.0.1-SNAPSHOT \
  --function.uri=file:classes?io.sk8s.invoker.java.function.FluxDoubler \
  --function.runner.isolated=true
```
