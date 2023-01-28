FROM circleci/golang

### Install AWS SAM ###
# Retrieve Shaws from: https://github.com/aws/aws-sam-cli/releases
ENV EXPECTED_SHAW="82d89dad1cef99ad34f334a3348c2a41137a2392b7e70c4a67075447c919d88d"
ENV FILE_NAME="aws-sam-cli-linux-x86_64.zip"
RUN wget "https://github.com/aws/aws-sam-cli/releases/latest/download/$FILE_NAME"
RUN sha256sum $FILE_NAME | sha256sum --check --status
RUN echo "$EXPECTED_SHAW $FILE_NAME" | sha256sum --check
RUN unzip $FILE_NAME -d sam-installation
RUN sudo ./sam-installation/install
RUN sam --version

RUN echo hello