FROM golang:1.11.1

ARG PROJECT_PATH
ENV GOPATH=/opt

RUN mkdir -p /opt/src/${PROJECT_PATH}

COPY . /opt/src/${PROJECT_PATH}/

WORKDIR /opt/src/${PROJECT_PATH}
RUN make configure

VOLUME /opt/src/${PROJECT_PATH}

ENTRYPOINT ["make", "run"]