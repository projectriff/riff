FROM java:8-alpine
VOLUME /tmp
COPY target/dependency/BOOT-INF/lib /app/lib
COPY target/dependency/META-INF /app/META-INF
COPY target/dependency/BOOT-INF/classes /app
ENTRYPOINT ["java","-Xmx128m","-Djava.security.egd=file:/dev/./urandom","-XX:TieredStopAtLevel=1","-noverify","-cp","app:app/lib/*","io.sk8s.invoker.java.server.JavaFunctionInvokerApplication"]
