FROM photon:4.0
COPY ./bin/idsreplay /idsreplay/
COPY emerging-all.rules /idsreplay/
CMD /idsreplay/idsreplay $IDSREPLAYOPTS
