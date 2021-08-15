FROM ubuntu

ARG APP_NAME
ENV EXEC="${APP_NAME}"

COPY bin/$EXEC /bin/$EXEC

COPY server.crt /
COPY server.key /

ENV GRPC_VERBOSITY=debug
ENV GRPC_TRACE=tcp,http,api

CMD ["/bin/$EXEC"]