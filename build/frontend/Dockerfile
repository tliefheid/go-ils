FROM golang:1.24-alpine

WORKDIR /app
COPY . .

ENV GO111MODULE=on
ENV CGO_ENABLED=0

RUN go generate ./...
RUN go build -mod=vendor -ldflags="-s -w" -o /usr/local/bin/service ./cmd/backend

FROM scratch
COPY --from=build /usr/local/bin/service /usr/local/bin/app
EXPOSE 8080/tcp

ENTRYPOINT ["/usr/local/bin/app"]
