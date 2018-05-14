FROM alpine:3.7

ARG GIT_COMMIT
ENV GIT_COMMIT ${GIT_COMMIT}

RUN apk --update upgrade && apk add ca-certificates && update-ca-certificates
COPY ./prometurbo.linux /bin/prometurbo

ENTRYPOINT ["/bin/prometurbo"]
