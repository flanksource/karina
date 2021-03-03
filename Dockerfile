FROM golang:1.13.6 as builder
WORKDIR /app
COPY ./ ./
ARG NAME
ARG VERSION
# upx 3.95 has issues compressing darwin binaries - https://github.com/upx/upx/issues/301
RUN  apt-get update && apt-get install -y xz-utils && \
  wget -nv -O upx.tar.xz https://github.com/upx/upx/releases/download/v3.96/upx-3.96-amd64_linux.tar.xz; tar xf upx.tar.xz; mv upx-3.96-amd64_linux/upx /usr/bin
RUN GOOS=linux GOARCH=amd64 make linux compress


FROM ubuntu:bionic
COPY --from=builder /app/.bin/karina /bin/
ARG SYSTOOLS_VERSION=3.6

RUN apt-get update && \
  apt-get install -y  genisoimage gnupg-agent curl apt-transport-https wget jq git sudo python-setuptools python-pip python-dev build-essential xz-utils ca-certificates unzip zip software-properties-common && \
  rm -Rf /var/lib/apt/lists/*  && \
  rm -Rf /usr/share/doc && rm -Rf /usr/share/man  && \
  apt-get clean

RUN wget -nv --no-check-certificate https://github.com/moshloop/systools/releases/download/${SYSTOOLS_VERSION}/systools.deb && dpkg -i systools.deb
ARG SOPS_VERSION=3.5.0
RUN install_deb https://github.com/mozilla/sops/releases/download/v${SOPS_VERSION}/sops_${SOPS_VERSION}_amd64.deb
RUN install_bin https://github.com/CrunchyData/postgres-operator/releases/download/v4.1.0/expenv
RUN install_bin https://github.com/hongkailiu/gojsontoyaml/releases/download/e8bd32d/gojsontoyaml
RUN pip install awscli

ENTRYPOINT [ "/bin/karina" ]
