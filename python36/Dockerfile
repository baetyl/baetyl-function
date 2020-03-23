FROM --platform=$TARGETPLATFORM baetyltechtest/python:3.6-devel
COPY *.py /bin/
RUN cd /bin/ && chmod +x runtime.py
ENTRYPOINT ["runtime.py"]