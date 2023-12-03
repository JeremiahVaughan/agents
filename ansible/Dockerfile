# Start from the Alpine base image
FROM alpine:latest

# Update the package list
RUN apk update \
    && apk add --no-cache python3 py3-pip \
    pip3 install ansible

RUN ansible --version

ENTRYPOINT ["/bin/sh", "-c"]
# The default command to run when starting the container
#CMD ["sleep", "3000000"]
