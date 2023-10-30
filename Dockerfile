FROM golang:1.21 as build

WORKDIR /go/src/s3-wrapper
COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o /go/bin/s3-wrapper

FROM gcr.io/distroless/static-debian12

COPY --from=build /go/bin/s3-wrapper /
CMD ["/s3-wrapper"]

