FROM golang:1.24-bookworm AS builder
LABEL authors="andrei"

WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/igo

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y git ca-certificates bash curl sudo

# Create tester user
RUN useradd -m -d /home/tester tester
RUN mkdir -p /home/tester/go/downloads /home/tester/go/versions /home/tester/go/bin /home/tester/go/shims /home/tester/go/scripts
RUN chown -R tester:tester /home/tester

COPY --from=builder /bin/igo /bin/igo
COPY tester.sh /home/tester/tester.sh
RUN chmod +x /bin/igo /home/tester/tester.sh

USER tester
WORKDIR /home/tester

ENV HOME=/home/tester

CMD ["-version"]