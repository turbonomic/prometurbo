FROM alpine:3.7

COPY ./icon/favicon.jpg /tmp/
COPY ./_output/appMetric.linux /bin/appMetric
RUN chmod +x /bin/appMetric
EXPOSE 8081

ENTRYPOINT ["/bin/appMetric"]
