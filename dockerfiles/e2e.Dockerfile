FROM golang:1.21-bookworm AS builder

WORKDIR /src/
COPY go.mod ./
RUN go mod download

COPY . .


RUN make install


FROM ubuntu:23.04

COPY --from=builder /go/bin/seda-chaind /usr/local/bin/

EXPOSE 26656 26657 1317 9090
CMD ["seda-chaind", "start"]
