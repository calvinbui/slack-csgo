FROM golang:alpine AS builder
WORKDIR /output
COPY . .
RUN go get -d -v
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /slack-csgo
RUN chmod +x /slack-csgo

FROM scratch
COPY --from=builder /slack-csgo /slack-csgo
ENTRYPOINT ["/slack-csgo"]
