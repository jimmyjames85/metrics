FROM golang:1.11 as builder

COPY . /go/src/github.com/jimmyjames85/metrics
WORKDIR /go/src/github.com/jimmyjames85/metrics
RUN CGO_ENABLED=0 GOOS=linux go build -a -o webservice cmd/webservice/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -o webclient cmd/webclient/main.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates curl jq bash
RUN adduser -D webservice
USER webservice

WORKDIR /webservice/

COPY --from=builder /go/src/github.com/jimmyjames85/metrics/webservice .
COPY --from=builder /go/src/github.com/jimmyjames85/metrics/webclient .

CMD ["/webservice/webservice"]
