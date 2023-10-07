FROM golang:1.21-alpine AS build-env

# Install minimum necessary dependencies
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev
RUN apk add --no-cache $PACKAGES


WORKDIR /go/src/github.com/sedaprotocol/seda-chain

# Optimized fetching of dependencies
COPY go.mod go.sum ./

RUN go mod download

# Copy and build the project
COPY . .

# Dockerfile Cross-Compilation Guide
# https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide
ARG TARGETOS TARGETARCH


RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build


FROM alpine:3

RUN apk add --no-cache curl make bash jq sed
COPY --from=build-env /go/src/github.com/cosmos/cosmos-sdk/build/seda-chaid /usr/bin/seda-chaind

EXPOSE 26656 26657 1317 9090

CMD ["seda-chaind", "start"]
STOPSIGNAL SIGTERM
WORKDIR /root
