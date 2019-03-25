FROM debian:stretch-slim

COPY accesslog-exporter /bin/accesslog-exporter

RUN mkdir -p /etc/accesslog_exporter
COPY etc/config.yaml /etc/accesslog_exporter/config.yaml
COPY etc/regexes.yaml /etc/accesslog_exporter/regexes.yaml

ENTRYPOINT [ "/bin/accesslog-exporter" ]

CMD [ "-config.path=/etc/accesslog_exporter/config.yaml", "-ua-regex.path=/etc/accesslog_exporter/regexes.yaml" ]
