FROM golang:1.18-alpine3.18 as build
WORKDIR /go/src/app
COPY ./go.mod /go/src/app/

RUN go mod download
RUN apk add --no-cache curl

# geoip db dl, it is free, but you need to grab a key
ARG MM_LICENSE_KEY=""
RUN if ! [ -z "$MM_LICENSE_KEY" ];then \
    curl -L "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=$MM_LICENSE_KEY&suffix=tar.gz" -o GeoLite2-City.tar.gz; \
    tar -xzf GeoLite2-City.tar.gz; \
    cp GeoLite2-City_*/GeoLite2-City.mmdb ./GeoLite2-City.mmdb; \
    else \
    touch GeoLite2-City.mmdb; \
    fi
ADD ./ /go/src/app/
RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -o /go/bin/app github.com/jothflee/honeygogo

# copy into our final image.
# use alpine since we need to hack in the entrypoint until we golang it
FROM alpine:3.18
RUN apk add --no-cache curl

COPY --from=build /go/bin/app /
COPY --from=build /go/src/app/GeoLite2-City.mmdb /
COPY ./entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app"]