FROM golang:1.17-alpine as build
WORKDIR /app
ADD . .
ENV CGO_ENABLED=0
RUN go test -v ./...
RUN go build -a -tags netgo -ldflags "-buildid='' -s -w" -trimpath -o cloudflare-cache-purger

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/cloudflare-cache-purger /
ENTRYPOINT ["/cloudflare-cache-purger"]
