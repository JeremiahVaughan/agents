FROM public.ecr.aws/docker/library/alpine:3.18.4 as builder

ARG TARGETPLATFORM
ARG ARCH


## Install 1password
RUN case "${TARGETPLATFORM}" in \
      "linux/amd64") ARCH="amd64" ;; \
      "linux/arm64") ARCH="arm64" ;; \
      "linux/386") ARCH="386" ;; \
      "linux/arm/v7") ARCH="arm" ;; \
      *) echo "Unsupported platform: ${TARGETPLATFORM}"; exit 1 ;; \
    esac && \
    wget "https://cache.agilebits.com/dist/1P/op2/pkg/v2.20.0/op_linux_${ARCH}_v2.20.0.zip" -O op.zip && \
    unzip -d op op.zip && \
    mv op/op /usr/bin/op


FROM public.ecr.aws/docker/library/alpine:3.18.4
COPY --from=builder /usr/bin/op /usr/bin/op

# Install aws-cli -- more than just a binary, also bash
RUN apk update && apk add --no-cache aws-cli bash

CMD ["/bin/bash"]