# Start from the Alpine base image
FROM alpine:latest

# Update the package list
RUN apk update \
    # Install ca-certificates for SSL verification during the download
    && apk add --no-cache ca-certificates \
    # Install wget to download the cloudflared binary
    && apk add --no-cache wget

# Download the cloudflared binary
RUN wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -O /usr/local/bin/cloudflared \
    # Make the binary executable
    && chmod +x /usr/local/bin/cloudflared

# Check if cloudflared is installed
RUN cloudflared --version

ENTRY_POINT ["/bin/sh", "-c"]
# The default command to run when starting the container
#CMD ["sleep", "3000000"]
