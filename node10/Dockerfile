FROM --platform=$TARGETPLATFORM node:10.19.0-buster-slim
COPY *.js *.json /bin/
RUN cd /bin/ && chmod +x runtime.js && npm install
ENTRYPOINT ["runtime.js"]