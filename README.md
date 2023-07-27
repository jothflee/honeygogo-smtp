# honeygogo-smtp standalone honeypot

A lightweight SMTP honeypot server written in Go, leveraging [go-smtp](github.com/emersion/go-smtp). A stand alone version of a module from honeygogo "a golang honeypot" ecosystem.

Logs mail to stdout and elasticsearch.

![crash gopher](https://github.com/jothflee/honeygogo-smtp/raw/main/docs/crash-dummy.png)  
_Goper image by [egonelbre](https://github.com/egonelbre/gophers)_

### honygogo-smtp logs metadata about the emails sent to it. It does not relay or save the email data.

# Usage Notes

- the smtp server runs on port 10025
- set the `ELASTICSEARCH_URL` env var with the elasticsearch url if desired (ex. `http://es:9200`)
- set the `MM_LICENSE_KEY` env var to pull maxmind geoip database ([this is a free thing](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data?lang=en))
- set the smtp server port with `HGG_PORT`

## [docker hub](https://hub.docker.com/r/jothflee/honeygogo-smtp)

```
docker pull jothflee/honeygogo-smtp
```

# Contributing/Development

### Getting started

run elasticsearch `9200` kibana `5601` and smtp `10025` using docker/docker-compose

```
docker-compose up
```

or

```
go mody tidy
go run main.go
```

### License

The software is using MIT License (MIT) - contributors welcome.
