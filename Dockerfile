FROM golang:1.20 AS builder
WORKDIR /github.com/iamucil/telebot

# Get the dependencies so it can be cached into a layer
COPY go.mod go.sum ./
RUN go mod download

# Now copy all the source
COPY . .

ARG token=

# ...and build it.
RUN CGO_ENABLED=0 go build -o ./bin/app \
  -ldflags="-s -w -X main.token=${token} -extldflags \"-static\"" \
  .

# build the runtime image
FROM alpine:3.11
WORKDIR /root/
RUN apk add --no-cache --virtual .build-deps \
  ca-certificates \
  && update-ca-certificates \
  # Clean up when done
  && rm -rf /tmp/* \
  && apk del .build-deps

COPY --from=builder /github.com/iamucil/telebot/bin/app ./tbot

EXPOSE 80 
EXPOSE 8080

ENTRYPOINT ["./tbot"]
