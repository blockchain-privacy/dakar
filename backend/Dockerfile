# build stage
FROM golang:1.16-alpine as build-stage
RUN mkdir /build
ADD . /build/
WORKDIR /build

# RUN go build -o crawler ./cmd/crawler/
RUN go build -o crawler ./cmd/crawler/

# production stage
FROM scratch
COPY --from=build-stage /build/crawler /app/
WORKDIR /app
CMD ["./crawler"]
