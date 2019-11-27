FROM golang:1.11

RUN apt-get update && apt-get install -y supervisor

WORKDIR /go/src/github.com/jimmyjames85/metrics

CMD ["supervisord", "-n"]
