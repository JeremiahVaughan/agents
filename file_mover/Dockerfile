# Start from the Alpine base image
FROM alpine:latest

# Update the package list
RUN apk update && \
    apk add openssh-client

ENTRYPOINT ["/bin/sh", "-c"]

#CMD ["sleep", "3000000"]
