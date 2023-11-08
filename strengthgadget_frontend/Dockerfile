FROM alpine/terragrunt as builder

FROM cimg/node:current
COPY --from=builder /usr/local/bin/terragrunt /usr/bin/terragrunt
COPY --from=builder /bin/terraform /usr/bin/terragrunt
CMD ["sleep", "30000"]