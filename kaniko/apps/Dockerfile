# builder image
FROM golang:1.14 as builder
LABEL maintainer="mav-MWP-Engg-All@mavenir.com"
WORKDIR /apps
COPY . .
RUN cd apps && CGO_ENABLED=0 go build

# final image
FROM alpine:latest
WORKDIR /apps
COPY --from=builder /apps .
RUN apk add curl && curl -LO https://get.helm.sh/helm-v3.4.2-linux-amd64.tar.gz
RUN tar -zxvf helm-v3.4.2-linux-amd64.tar.gz && mv linux-amd64/helm /usr/local/bin/helm
RUN apk add --update docker openrc
RUN rc-update add docker boot
CMD ["sleep", "1200"]
