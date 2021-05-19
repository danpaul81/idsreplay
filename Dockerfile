FROM photon:latest
COPY ./idsreplay /idsreplay/
COPY emerging-all.rules /idsreplay/
CMD /idsreplay/idsreplay $IDSREPLAYOPTS
