ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /plausible-proxy .

FROM gcr.io/distroless/cc

EXPOSE 8080

USER nobody

COPY --from=builder /plausible-proxy /usr/local/bin/

CMD ["plausible-proxy"]
