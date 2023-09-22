# build stage
FROM golang:1.21-alpine as build-stage
ARG BLOCKCHAIN_MODE="unkown"
ARG MODFILE_PATH="go.mod"

WORKDIR /build
COPY . /build/

RUN go build -v -ldflags "-X main.blockchainMode=$BLOCKCHAIN_MODE" -modfile $MODFILE_PATH -o crawler ./cmd/crawler/

# production stage
FROM scratch
COPY --from=build-stage /build/crawler /app/

VOLUME ["/data"]
WORKDIR /app
ENTRYPOINT ["./crawler", "--config", "/data/config.yml"]
