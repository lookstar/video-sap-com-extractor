FROM ubuntu
LABEL owner=cosine.yan@sap.com
COPY extractor /opt/sap/
ENTRYPOINT ["/opt/sap/extractor"]