FROM golang:alpine3.12 as builder

RUN mkdir /app
RUN chmod 0777 /app

COPY main.go /app
COPY go.mod /app
COPY go.sum /app

WORKDIR /app

# import golang packages to be used inside image "scratch"
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/main /app/main.go

FROM scratch

COPY --from=builder /app/main main
VOLUME /scan
CMD ["./main"]