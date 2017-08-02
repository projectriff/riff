FROM openjdk:8-jdk-alpine
ADD target/topic-controller-0.0.1-SNAPSHOT.jar topic-controller.jar
ENTRYPOINT [ "java", "-Djava.security.egd=file:/dev/./urandom", "-jar", "topic-controller.jar" ]
EXPOSE 8080
