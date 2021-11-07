# honeygogo-smtp standalone honeypot

A lightweight SMTP honeypot server written in Go, leveraging [go-smtp](github.com/emersion/go-smtp). A stand alone version of a module from honeygogo "a golang honeypot" ecosystem.

Logs mail to stdout and elasticsearch.

![crash gopher](./docs/crash-dummy.png)  
_Goper image by [egonelbre](https://github.com/egonelbre/gophers)_

### honygogo-smtp logs metadata about the emails sent to it. It does not relay or save the email data.

# Usage

the smtp server runs on port 10025  
set the `ELASTICSEARCH_URL` env var with the elasticsearch url if desired (ex. `http://es:9200`)

# Contributing/Development

### Getting started

```
go mody tidy
go run main.go
```

### License

The software is using MIT License (MIT) - contributors welcome.
