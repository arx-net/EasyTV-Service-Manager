#
#   Building the executable
#
FROM golang:1.11.3-stretch as builder

RUN go get golang.org/x/crypto/bcrypt

RUN go get github.com/lib/pq

RUN go get github.com/sirupsen/logrus

COPY ./src /go/src/

WORKDIR /go/src/gitlab.arx.net/easytv/sm/cmd/cron_job

RUN CGO_ENABLED=0 GOOS=linux go install

#
#   Creating the final image
#
FROM alpine:3.8


COPY --from=builder /go/bin/cron_job /app/

RUN chmod +x /app/cron_job

RUN mkdir -p /var/log/sm

COPY ./crontab.txt .

RUN crontab crontab.txt

CMD [ "crond" , "-f"]