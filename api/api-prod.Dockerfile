FROM golang:1.11.3-stretch as builder

# Install the dependencies
RUN go get -u github.com/go-chi/chi

RUN go get github.com/lib/pq

RUN go get github.com/bradfitz/gomemcache/memcache

RUN go get golang.org/x/crypto/bcrypt

RUN go get github.com/sirupsen/logrus

RUN go get gopkg.in/urfave/cli.v1

RUN go get github.com/olekukonko/tablewriter

# Copy all the code
COPY ./src /go/src/

# Build the srt
WORKDIR /go/src/gitlab.arx.net/easytv/sm/cmd/srt
RUN CGO_ENABLED=0 GOOS=linux go install

# Build the db tool
WORKDIR /go/src/gitlab.arx.net/easytv/sm/cmd/init_db
RUN CGO_ENABLED=0 GOOS=linux go install

# Build the api tool
WORKDIR /go/src/gitlab.arx.net/easytv/sm/cmd/api
RUN CGO_ENABLED=0 GOOS=linux go install

#
#   Minimal image
#
FROM alpine:latest

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=builder /go/bin/srt .

COPY --from=builder /go/bin/init_db .

COPY --from=builder /go/bin/api .

ENTRYPOINT ./api