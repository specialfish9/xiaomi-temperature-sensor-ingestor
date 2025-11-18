FROM golang:alpine as build-env

WORKDIR /go/src/app/
COPY ./ /go/src/app/ 

RUN go get -d -v ./...

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static
COPY --from=build-env /go/bin/app /
WORKDIR /
CMD ["/app"]

