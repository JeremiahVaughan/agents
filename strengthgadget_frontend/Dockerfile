FROM alpine/terragrunt as builder

RUN apk update && apk add --no-cache nodejs-current npm

#FROM cimg/node:current
#COPY --from=builder /usr/local/bin/terragrunt /usr/bin/terragrunt
#COPY --from=builder /bin/terraform /usr/bin/terragrunt
#CMD ["sleep", "30000"]