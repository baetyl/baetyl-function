FROM --platform=$TARGETPLATFORM baetyltechtest/node:10.19-devel
COPY *.js /bin/
RUN cd /bin/ && chmod +x runtime.js
ENTRYPOINT ["runtime.js"]