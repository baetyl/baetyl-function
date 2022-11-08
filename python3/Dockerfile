ARG PYTHON_VERSION=3.7
FROM --platform=$TARGETPLATFORM baetyltechtest/python:${PYTHON_VERSION}-devel
COPY *.py /bin/
RUN cd /bin/ && chmod +x runtime.py
ENTRYPOINT ["runtime.py"]
