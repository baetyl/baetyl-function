FROM --platform=$TARGETPLATFORM python:3-slim-buster
COPY *.py *.txt /bin/
RUN cd /bin/ && chmod +x runtime.py && pip install -r requirements.txt
ENTRYPOINT ["runtime.py"]