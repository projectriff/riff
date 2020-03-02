FROM oracle/graalvm-ce:19.2.0.1 AS native
RUN gu install native-image
COPY target/*.jar /tmp/app.jar
RUN mkdir /tmp/app && cd /tmp/app && jar -xf /tmp/app.jar
RUN mkdir /app
RUN mv /tmp/app/BOOT-INF/lib /app/lib
RUN mv /tmp/app/META-INF /app/META-INF
RUN cp -r /tmp/app/BOOT-INF/classes/* /app
RUN native-image -Dio.netty.noUnsafe=true --no-server -H:Name=app/main -H:+ReportExceptionStackTraces --no-fallback --allow-incomplete-classpath --report-unsupported-elements-at-runtime -DremoveUnusedAutoconfig=true -cp app:`echo app/lib/*.jar | tr ' ' :` io.projectriff.processor.Processor

FROM ubuntu:bionic
COPY --from=native /app/main /app/main
ENTRYPOINT ["/app/main","${0}","${@}"]
