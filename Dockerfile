FROM alpine:latest

RUN apk add --no-cache tzdata

ENV SLM_DATA_DIR /usr/local/share/switch-library-manager-web

RUN mkdir -p $SLM_DATA_DIR

COPY build/switch-library-manager-web /usr/local/bin/switch-library-manager-web

VOLUME $SLM_DATA_DIR
VOLUME /mnt/roms

EXPOSE 3000

CMD ["switch-library-manager-web"]
