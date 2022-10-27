FROM golang:1.17-alpine3.16

# Update
RUN apk update && apk upgrade --no-cache

# Install node, zip, git and make
RUN apk add --no-cache gcc libc-dev npm nodejs-current zip git openssh make

# Install aws cli
RUN apk add --no-cache \
    python3 \
    py3-pip \
    && pip3 install --upgrade pip \
    && pip3 install --no-cache-dir \
    awscli \
    && rm -rf /var/cache/apk/*

# Add cli binary
ADD . /app
WORKDIR /app

# Copy cli to bin location
RUN cp sailpoint-cli /usr/local/bin/sail