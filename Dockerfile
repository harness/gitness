FROM drone_golang

COPY . /go/src/github.com/drone/drone
WORKDIR /go/src/github.com/drone/drone

ENV GO15VENDOREXPERIMENT 1

RUN make gen_static && make build_static

ADD .env /.env

ENTRYPOINT ["./drone_static"]

EXPOSE 9898


