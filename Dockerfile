FROM photon:latest
COPY ./bin/idsreplay /idsreplay/
COPY emerging-all.rules /idsreplay/
CMD /idsreplay/idsreplay $IDSREPLAYOPTS
