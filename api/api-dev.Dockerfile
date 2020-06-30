FROM golang:1.11.3-stretch

RUN go get -u github.com/go-chi/chi

RUN go get github.com/lib/pq

RUN go get github.com/codegangsta/gin

RUN go get github.com/bradfitz/gomemcache/memcache

RUN go get golang.org/x/crypto/bcrypt

RUN go get github.com/sirupsen/logrus

RUN go get gopkg.in/urfave/cli.v1

RUN go get github.com/olekukonko/tablewriter

COPY ./src /go/src/

WORKDIR /go/src/gitlab.arx.net/easytv/sm/cmd/srt
RUN go install

WORKDIR /go/src/gitlab.arx.net/easytv/sm/cmd/init_db
RUN go install

WORKDIR /go/src/gitlab.arx.net/easytv/sm/cmd/api
RUN go install

CMD ["gin", "run", "api"]