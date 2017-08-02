FROM openjdk:8-jdk-alpine
ADD target/topic-gateway-0.0.1-SNAPSHOT.jar topic-gateway.jar
ENTRYPOINT [ "java", "-Djava.security.egd=file:/dev/./urandom", "-jar", "topic-gateway.jar" ]
EXPOSE 8080
