FROM --platform=$TARGETPLATFORM golang:1.13.5-stretch as devel
ARG BUILD_ARGS
COPY / /go/src/
RUN cd /go/src/ && make build BUILD_ARGS=$BUILD_ARGS

FROM --platform=$TARGETPLATFORM busybox
ARG TARGETPLATFORM
COPY --from=devel /go/src/output/$TARGETPLATFORM/baetyl-function/baetyl-function /bin/
ENTRYPOINT ["baetyl-function"]
