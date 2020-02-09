FROM golang:1.13 as build
WORKDIR /app
ADD . .
RUN CGO_ENABLED=0 go build -a -tags netgo -ldflags "-buildid='' -s -w" -trimpath -o cloudflare-cache-purger

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/cloudflare-cache-purger /
ENTRYPOINT ["/cloudflare-cache-purger"]