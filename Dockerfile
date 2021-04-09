FROM ubuntu

ARG APP_NAME
ENV EXEC="${APP_NAME}"

COPY bin/$EXEC /bin/$EXEC
COPY entrypoint.sh /entrypoint.sh

COPY server.crt /
COPY server.key /

ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]