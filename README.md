[![Build Status](https://travis-ci.org/vlamug/accesslog-exporter.svg?branch=master)](https://travis-ci.org/vlamug/accesslog-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/vlamug/accesslog-exporter)](https://goreportcard.com/report/github.com/vlamug/accesslog-exporter)

# Accesslog Exporter

The exporter accepts accesslog via syslog stream and exposes metrics readable for Prometheus.

## Requirements

The Go with version 1.11 should be installed.

## Build and run

To build and run exporter perform the following command:

```
$> make build
$> ./accesslog-exporter --config.path=etc/config.yaml --ua-regex.path=etc/regexes.yaml --web.addr=:9032 --syslog.addr=:9033
```

Where:
 - `config.path` - the path to the exporter config. Default value: `etc/config.yaml`.
 - `ua-regex.path` - the path to the file, that contains all regexes for parsing user agent. Default value: `etc/regexes.yaml`.
 - `web.addr` - the address on which the metrics are exposed. Default value: `:9032`.
 - `syslog.addr` - the address on which the syslog accept requests(Nginx access log lines). Default value: `:9033`.

## Tests

To run tests use the command:

```
make test
```

## Benchmarks

To run benchmarks use the command:

```
make bench
```

## Config explanation

This is an example of configuration that cover all aspects of Exporter features:

```yaml
global:
  # (optional) If log string contains $remote_addr that
  # belongs to following subnets - User Agent won't be parsed
  # Instead of that OS and Device will be marked as "internal"
  internal_subnets:
    - 30.0.0.0/8
    - 188.173.191.0/24

  # (optional) User agents cache max size (items). Default - 100k
  user_agent_cache_size: 200000

  # (optional) Number of workers(goroutines) to parse and export log line metrics. Default - 100
  export_workers: 1000

  # (optional) Use this for custom User Agent replacements
  user_agents:
    - match_re: ^MyStore\/([0-9]+)
      replacements:
        os: IOS
        user_agent: myapp_ios_$1
    - match_re: ^myapp_android\/([0-9\.]+)
      replacements:
        os: Android
        user_agent: myapp_android_$1 # $1 will be replaced by version

  # (optional) If defined - additional metric will be collected for particular path
  request_uris:
    - match_re: ^/(\?.*)?$
      match_method: "GET" # (optional)
      replacements:
        request_uri: home
    - match_re: ^/search/.*
      match_method: "GET" # (optional)
      replacements:
        request_uri: search

  # (optional) Contains list of equivalent hosts, that should considered as the same, for example: www.site.com, site.com
  hosts:
    - match: 'site.ru'
      replacement: 'www.site.ru'

# (required) List of your Nginx hosts to collect logs from
sources:
  - host: loadbalancer
    # Nginx log format.
    # See https://nginx.org/ru/docs/http/ngx_http_log_module.html
    # Following multiline string will be simple one-line string
    # See YAML syntax here:
    # https://symfony.com/doc/current/components/yaml/yaml_format.html
    log_format: >
      [$time_local] | $remote_addr | $remote_user | $status | $scheme | $host | "$request" | $body | $body_bytes_sent| $fullrequest | $http_user_agent | $http_referer | $request_time | $connection_requests
```

The config file consists of two sections:
1. `Global` -  contains filters, replacements, cache, worker settings.
2. `Sources` - contains list of Nginx hosts with access log formats. It should have at least one accesslog format!

Lets examine each parameter in `Global` section:

| parameter | required | default value | description |
|---|---|---|---|
| internal_subnets | no  | - | It is a list of subnets that are considered as internal and for which the requests should not be processed. |
| user_agent_cache_size | yes  | 100000 | Defines how many parsed user agent should be stored in memory. The cache allows to process the incoming logs stream faster. |
| export_workers | yes  | 100 | Number of workers(actually goroutines) to parse and export log lines. |
| user_agents | no | - | Is used for custom User Agent replacements in metrics labels. |
| request_uris | no | - | Is used to collect additional metric(`uri_response_time_seconds`) by particular uri path. |
| hosts | no | - | Contains list of equivalent hosts, that should considered as the same, for example: www.site.com and site.com. |
