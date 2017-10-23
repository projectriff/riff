FROM java:8-alpine
VOLUME /tmp
ARG JAR_FILE
ADD ./target/${JAR_FILE} /java-function-invoker.jar
ENTRYPOINT ["java","-Djava.security.egd=file:/dev/./urandom","-jar","/java-function-invoker.jar"]
