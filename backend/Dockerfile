# build stage
FROM golang:1.19-alpine as build-stage

WORKDIR /build
COPY . /build/

RUN go build -o crawler ./cmd/crawler/

# production stage
FROM scratch
COPY --from=build-stage /build/crawler /app/
WORKDIR /app
CMD ["./crawler"]
