FROM ubuntu

RUN apt-get update && apt-get install wget lsb-core unzip zip jq zsh -y


### Install Github CLI ###
# Get install instructions from https://github.com/cli/cli/blob/trunk/docs/install_linux.md
RUN type -p curl >/dev/null || apt-get install curl -y && \
    curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install gh -y


### Install Terraform ###
# Get hashes from https://releases.hashicorp.com/terraform/1.3.8/terraform_1.3.8_SHA256SUMS
ENV EXPECTED_SHAW="9d9e7d6a9b41cef8b837af688441d4fbbd84b503d24061d078ad662441c70240"
ENV TERRAFORM_VERSION="1.3.8"
ENV FILE_NAME="terraform_${TERRAFORM_VERSION}_linux_amd64.zip"
RUN wget "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/${FILE_NAME}"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN unzip $FILE_NAME -d terraform
RUN mv terraform/terraform /usr/bin/
RUN terraform --version


### Install Terragrunt
# Get hashes from https://github.com/gruntwork-io/terragrunt/releases
ENV EXPECTED_SHAW="d9886c7de21fbe7127251735599f634eb7899f635e7cd73015a7b1a0a25772fc"
ENV TERRAGRUNT_VERSION="0.43.2"
ENV FILE_NAME="terragrunt_linux_amd64"
RUN wget "https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN chmod 700 $FILE_NAME
RUN mv $FILE_NAME /usr/bin/terragrunt
RUN terragrunt --version


### Install Golang ###
# Get hashes from https://go.dev/dl/
ENV EXPECTED_SHAW="5a9ebcc65c1cce56e0d2dc616aff4c4cedcfbda8cc6f0288cc08cda3b18dcbf1"
ENV GOLANG_VERSION="1.20"
ENV FILE_NAME="go${GOLANG_VERSION}.linux-amd64.tar.gz"
RUN wget "https://go.dev/dl/go1.20.linux-amd64.tar.gz"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf "go${GOLANG_VERSION}.linux-amd64.tar.gz"
ENV PATH=$PATH:/usr/local/go/bin
RUN go version
# Needed for pipeline script
# todo get this working
#RUN go install github.com/aws/aws-sdk-go@latest


### Install Node ###
# Get hashes from https://nodejs.org/en/download/
ENV EXPECTED_SHAW="1ccec74b6240fce8754813e31fdbc93ad520df2e814729cea29efe9075e48350"
ENV NODE_VERSION="18.14.0"
ENV FILE_NAME="node-v${NODE_VERSION}-linux-x64.tar.xz"
RUN wget "https://nodejs.org/dist/v${NODE_VERSION}/${FILE_NAME}"
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN tar -xf $FILE_NAME
ENV PATH=$PATH:"/node-v${NODE_VERSION}-linux-x64/bin"
RUN node -v


RUN mkdir -p /workspace
WORKDIR /workspace
COPY aws-cli-integrity-key .
ENV GNUGP_KEY_FILE="aws-cli-integrity-key"
ENV FILE_NAME="awscli-exe-linux-x86_64.zip"
ENV FILE_SIGNATURE_NAME="$FILE_NAME.sig"
# Get hashes from https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html
RUN wget "https://awscli.amazonaws.com/$FILE_NAME"
RUN wget "https://awscli.amazonaws.com/$FILE_SIGNATURE_NAME"
RUN gpg --import $GNUGP_KEY_FILE
RUN gpg --verify $FILE_SIGNATURE_NAME $FILE_NAME
RUN unzip $FILE_NAME
RUN ./aws/install
RUN aws --version
