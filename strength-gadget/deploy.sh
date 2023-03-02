#!/bin/zsh
set -e

#gh gist clone https://github.com/JeremiahVaughan/strength-gadget-v3.git

# todo once you get the hash checks working, look into getting rid of the setup workflow as its probably going to be redundant
go run "$HOME/strength-gadget-v3/deploy.go"
cd "$HOME/strength-gadget-v3/infrastructure/terraform/live/$ENVIRONMENT/lambdas" && \
    terragrunt init \
      -backend-config="access_key=$TERRAFORM_STATE_BUCKET_KEY" \
      -backend-config="secret_key=$TERRAFORM_STATE_BUCKET_SECRET" \
      -backend-config="region=$TERRAFORM_STATE_BUCKET_REGION" && \
    terragrunt apply -auto-approve

aws ssm get-parameter --name lambda_urls --query Parameter.Value --output text > "$HOME/strength-gadget-v3/ui-react-2/src/assets/env.json"

cd "$HOME/strength-gadget-v3/ui-react-2"
# todo will want to find a way to cache this install to speed up deployments. Maybe CircleCI docker cache can be leveraged.
npm i
npx nx build --configuration=production

cd "$HOME/strength-gadget-v3/infrastructure/terraform/live/$ENVIRONMENT/cloudfront" && \
    terragrunt init \
      -backend-config="access_key=$TERRAFORM_STATE_BUCKET_KEY" \
      -backend-config="secret_key=$TERRAFORM_STATE_BUCKET_SECRET" \
      -backend-config="region=$TERRAFORM_STATE_BUCKET_REGION" && \
    terragrunt apply -auto-approve

# todo add cloudflare in front of cloudfront for better caching and security
# todo add cloudflare in front of cloudfront for better caching and security
