FROM golang:latest AS build-stage

ENV CGO_ENABLED=0

# download dependencies before build, so they can be cached
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o /dakar ./cmd/dakar/

FROM scratch
COPY --from=build-stage /dakar .

VOLUME ["/data"]
CMD ["--config", "/data/config.yml"]
ENTRYPOINT ["/dakar"]
