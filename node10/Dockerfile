FROM --platform=$TARGETPLATFORM hub.baidubce.com/baetyl/node:10.23-devel
COPY runtime.js function_pb.js function_grpc_pb.js /bin/
RUN cd /bin/ && chmod +x runtime.js
ENTRYPOINT ["runtime.js"]