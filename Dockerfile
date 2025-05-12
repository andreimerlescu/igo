FROM golang:1.24-bookworm AS builder
LABEL authors="andrei"

WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/igo

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y git ca-certificates bash curl sudo tree bc

# Create tester user
RUN useradd -m -d /home/tester tester
RUN chown -R tester:tester /home/tester

COPY --from=builder /bin/igo /bin/igo
ENV PATH=/bin
COPY tester.sh /home/tester/tester.sh
RUN chmod +x /bin/igo /home/tester/tester.sh
USER tester
WORKDIR /home/tester

ENV HOME=/home/tester

CMD ["-version"]