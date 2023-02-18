FROM ubuntu

RUN apt-get update && apt-get install wget lsb-core unzip zip jq -y


### Install Terraform ###
ENV EXPECTED_SHAW="9d9e7d6a9b41cef8b837af688441d4fbbd84b503d24061d078ad662441c70240"
ENV TERRAFORM_VERSION="1.3.8"
ENV FILE_NAME="terraform_${TERRAFORM_VERSION}_linux_amd64.zip"
# Get hashes from https://releases.hashicorp.com/terraform/1.3.8/terraform_1.3.8_SHA256SUMS
RUN wget "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/${FILE_NAME}"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN unzip $FILE_NAME -d terraform
RUN mv terraform/terraform /usr/bin/
RUN terraform --version


### Install Terragrunt
ENV EXPECTED_SHAW="d9886c7de21fbe7127251735599f634eb7899f635e7cd73015a7b1a0a25772fc"
ENV TERRAGRUNT_VERSION="0.43.2"
ENV FILE_NAME="terragrunt_linux_amd64"
# Get hashes from https://github.com/gruntwork-io/terragrunt/releases
RUN wget "https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN chmod 700 $FILE_NAME
RUN mv $FILE_NAME /usr/bin/terragrunt
RUN terragrunt --version


### Install Golang ###
ENV EXPECTED_SHAW="5a9ebcc65c1cce56e0d2dc616aff4c4cedcfbda8cc6f0288cc08cda3b18dcbf1"
ENV GOLANG_VERSION="1.20"
ENV FILE_NAME="go${GOLANG_VERSION}.linux-amd64.tar.gz"
# Get hashes from https://go.dev/dl/
RUN wget "https://go.dev/dl/go1.20.linux-amd64.tar.gz"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf "go${GOLANG_VERSION}.linux-amd64.tar.gz"
ENV PATH=$PATH:/usr/local/go/bin
RUN go version


ENV EXPECTED_SHAW="1ccec74b6240fce8754813e31fdbc93ad520df2e814729cea29efe9075e48350"
ENV NODE_VERSION="18.14.0"
ENV FILE_NAME="node-v${NODE_VERSION}-linux-x64.tar.xz"
# Get hashes from https://nodejs.org/en/download/
### Install Node ###
RUN wget "https://nodejs.org/dist/v${NODE_VERSION}/${FILE_NAME}"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN tar -xf $FILE_NAME
ENV PATH=$PATH:"/node-v${NODE_VERSION}-linux-x64/bin"
RUN node -v
