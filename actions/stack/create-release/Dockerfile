FROM alpine

RUN apk add \
    bash \
    curl \
    git \
  && rm -rf /var/cache/apk/*

ARG bosh_version=7.1.3

RUN mkdir -p "/bin" \
      && export PATH="${PATH}:/bin" \
      && curl "https://github.com/cloudfoundry/bosh-cli/releases/download/v${bosh_version}/bosh-cli-${bosh_version}-linux-amd64" \
        --fail-with-body \
        --show-error \
        --silent \
        --location \
        --output "/bin/bosh" \
      && chmod +x "/bin/bosh"

COPY entrypoint /entrypoint
ENTRYPOINT ["/entrypoint"]
