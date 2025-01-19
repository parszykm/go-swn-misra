FROM golang:1.23-alpine as builder

WORKDIR /app

COPY go.mod ./

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/misra

FROM alpine:latest

COPY --from=builder /app/bin/misra /bin/misra

ENTRYPOINT ["misra"]
CMD ["--help"]




