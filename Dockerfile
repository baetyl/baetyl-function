FROM --platform=$TARGETPLATFORM golang:1.13.5-stretch as devel
COPY / /go/src/
RUN cd /go/src/ && make all

FROM --platform=$TARGETPLATFORM busybox
ARG TARGETPLATFORM
COPY --from=devel /go/src/output/$TARGETPLATFORM/baetyl-function/baetyl-function /bin/
ENTRYPOINT ["baetyl-function"]
