FROM golang:latest AS build-stage

# download dependencies before build, so they can be cached
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o /crawler ./cmd/crawler/

FROM scratch
COPY --from=build-stage /crawler /app/

VOLUME ["/data"]
WORKDIR /app
CMD ["--config", "/data/config.yml"]
ENTRYPOINT ["./crawler"]
