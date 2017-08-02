FROM openjdk:8-jdk-alpine
ADD target/function-dispatcher-0.0.1-SNAPSHOT.jar function-dispatcher.jar
ENTRYPOINT [ "java", "-Djava.security.egd=file:/dev/./urandom", "-jar", "function-dispatcher.jar" ]
EXPOSE 8080
