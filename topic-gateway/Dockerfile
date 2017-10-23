FROM java:8-alpine
VOLUME /tmp
ARG JAR_FILE
ADD ./target/${JAR_FILE} /topic-gateway.jar
ENTRYPOINT ["java","-Djava.security.egd=file:/dev/./urandom","-jar","/topic-gateway.jar"]
