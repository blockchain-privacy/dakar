# build stage
FROM golang:1.16-alpine as build-stage
ENV GO111MODULE='on'
RUN mkdir /build
ADD . /build/
WORKDIR /build

# RUN go build -o crawler ./cmd/crawler/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o crawler ./cmd/crawler/

# production stage
FROM scratch
COPY --from=build-stage /build/crawler /app/
WORKDIR /app
CMD ["./crawler", "-dash", "-dbhost", "alpha", "-rpchost", "dashd"]
