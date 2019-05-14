FROM golang

WORKDIR /go/src/slack_emoji_notify
COPY . .

ENV GO111MODULE on
RUN go get -u -v
RUN go install -v ./...

CMD ["slack_emoji_notify"]
