FROM ubuntu

COPY backend/storage /

COPY server.crt /
COPY server.key /

EXPOSE 8080

# Command to run
ENTRYPOINT ["/backend/storage"]
