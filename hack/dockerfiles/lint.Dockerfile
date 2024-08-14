#syntax=docker/dockerfile:1.8
#check=error=true

ARG GO_VERSION=1.22
ARG ALPINE_VERSION=3.20
ARG XX_VERSION=1.4.0
ARG GOLANGCI_LINT_VERSION=1.60.1

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS golang-base
FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM golang-base AS base
ENV GOFLAGS="-buildvcs=false"
ARG GOLANGCI_LINT_VERSION
RUN apk add --no-cache yamllint
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v${GOLANGCI_LINT_VERSION}
COPY --link --from=xx / /
WORKDIR /go/src/github.com/tonistiigi/go-csvvalue

FROM base AS golangci-lint
ARG BUILDTAGS
ARG TARGETPLATFORM
RUN --mount=type=bind \
    --mount=target=/root/.cache,type=cache \
  xx-go --wrap && \
  golangci-lint run --build-tags "${BUILDTAGS}" && \
  touch /golangci-lint.done
  
FROM base AS golangci-verify
RUN --mount=type=bind \
  golangci-lint config verify && \
  touch /golangci-verify.done

FROM base AS yamllint
RUN --mount=type=bind \
  yamllint -c .yamllint.yml --strict . && \
  touch /yamllint.done  

FROM scratch
COPY --link --from=golangci-lint /golangci-lint.done /
COPY --link --from=yamllint /yamllint.done /
COPY --link --from=golangci-verify /golangci-verify.done /