FROM golang:1.24-alpine AS build


COPY . .
RUN go generate ./...

ENV GO111MODULE=on
ENV CGO_ENABLED=0

RUN go build -mod=vendor -ldflags="-s -w" -o /service ./cmd/frontend

FROM scratch
COPY --from=build /service /app
COPY cmd/frontend/assets /assets/
EXPOSE 3000/tcp


ENTRYPOINT ["/app"]
