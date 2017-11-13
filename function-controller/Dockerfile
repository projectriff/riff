FROM java:8-alpine
VOLUME /tmp
ARG JAR_FILE
ADD ./target/${JAR_FILE} /function-controller.jar
ENTRYPOINT ["java","-Djava.security.egd=file:/dev/./urandom","-jar","/function-controller.jar"]
