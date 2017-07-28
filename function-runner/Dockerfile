FROM openjdk:8-jdk-alpine
ADD target/function-runner-0.0.1-SNAPSHOT.jar function-runner.jar
ENTRYPOINT [ "java", "-Djava.security.egd=file:/dev/./urandom", "-jar", "function-runner.jar" ]
EXPOSE 8080

