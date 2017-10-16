FROM python:3.6-slim
VOLUME /tmp
ADD ./java/greeter/target/greeter-1.0.0.jar /functions/java/greeter-1.0.0.jar
ADD ./node/hello.js /functions/node/hello.js
ADD ./script/wordcount.sh /functions/script/wordcount.sh
RUN sh -c 'touch /functions/java/greeter-1.0.0.jar'
RUN sh -c 'touch /functions/node/hello.js'
RUN sh -c 'touch /functions/script/wordcount.sh'
CMD cd functions && python -m http.server 8080
