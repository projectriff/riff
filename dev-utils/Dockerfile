# build stage
FROM golang:1.13 AS build
ADD . /src
RUN cd /src \
  && go build ./cmd/subscribe \
  && go build ./cmd/publish


# final stage
FROM ubuntu:bionic

ADD scripts/* /riff/dev-utils/bin/

COPY --from=build /src/subscribe /riff/dev-utils/bin
COPY --from=build /src/publish /riff/dev-utils/bin

WORKDIR /riff/dev-utils

ENV PATH="/riff/dev-utils/bin:${PATH}"

RUN apt-get update \
  && apt-get install -y bash-completion \
  && mkdir -p /etc/bash_completion.d \
  && echo ". /etc/profile.d/bash_completion.sh" >> ~/.bashrc \
  && apt-get install -y curl \
  && apt-get install -y gnupg2 \
  && apt-get install -y apt-transport-https \
  && curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - \
  && echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list \
  && apt-get update \
  && apt-get install -y kubectl \
  && kubectl completion bash > /etc/bash_completion.d/kubectl \
  && apt-get install -y jq \
  && apt-get remove -y --auto-remove apt-transport-https \
  && apt-get remove -y --auto-remove gnupg2 \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/* \
  && curl -L https://storage.googleapis.com/projectriff/riff-cli/releases/v0.5.0-snapshot/riff-linux-amd64.tgz | tar xz -C /riff/dev-utils/bin \
  && riff completion --shell bash > /etc/bash_completion.d/riff

CMD ["entrypoint.sh"]
