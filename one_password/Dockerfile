FROM public.ecr.aws/docker/library/alpine:3.18.4 as builder

# Install 1password
RUN echo https://downloads.1password.com/linux/alpinelinux/stable/ >> /etc/apk/repositories && \
    wget https://downloads.1password.com/linux/keys/alpinelinux/support@1password.com-61ddfc31.rsa.pub -P /etc/apk/keys && \
    apk update && apk add 1password-cli


FROM public.ecr.aws/docker/library/alpine:3.18.4
COPY --from=builder /usr/bin/op /usr/bin/op

# Install aws-cli -- more than just a binary, also bash
RUN apk update && apk add --no-cache aws-cli bash

CMD ["/bin/bash"]