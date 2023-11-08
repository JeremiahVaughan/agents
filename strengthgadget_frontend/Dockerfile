FROM alpine/terragrunt as builder
RUN apk add --update util-linux

FROM cimg/base:current
COPY --from=builder /usr/local/bin/terragrunt /usr/bin/terragrunt
COPY --from=builder /bin/terraform /usr/bin/terragrunt
