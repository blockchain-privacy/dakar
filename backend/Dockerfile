# build stage
FROM golang:1.22-alpine as build-stage
ARG BLOCKCHAIN_MODE="unkown"
ARG MODFILE_PATH="go.mod"
ARG MODSUM_PATH="go.sum"

WORKDIR /build

# download dependencies before build, so they can be cached
COPY $MODFILE_PATH ./go.mod
COPY $MODSUM_PATH ./go.sum
RUN go mod download

COPY . .
RUN go build -v -ldflags "-X main.blockchainMode=$BLOCKCHAIN_MODE" -modfile $MODFILE_PATH -o crawler ./cmd/crawler/

# production stage
FROM scratch
COPY --from=build-stage /build/crawler /app/

VOLUME ["/data"]
WORKDIR /app
CMD ["--config", "/data/config.yml"]
ENTRYPOINT ["./crawler"]
