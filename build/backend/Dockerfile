FROM golang:1.24-alpine

WORKDIR /app
COPY . .

ENV GO111MODULE=on
ENV CGO_ENABLED=0

RUN go build -mod=vendor -ldflags="-s -w" -o /usr/local/bin/backend ./cmd/backend

FROM scratch
COPY --from=build /usr/local/bin/backend /usr/local/bin/app
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/app"]
