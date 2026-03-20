FROM golang:1.26-bookworm AS builder
LABEL authors="andrei"

WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /bin/igo

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y git ca-certificates bash curl sudo tree bc

RUN curl -sL "https://github.com/andreimerlescu/bump/releases/download/v1.0.4-beta.1/bump-linux-amd64" -o /bin/bump && \
    chmod +x /bin/bump

# Create tester user
RUN useradd -m -d /home/tester tester
RUN chown -R tester:tester /home/tester

COPY --from=builder /bin/igo /bin/igo
ENV PATH=/bin
COPY testing/*.sh /home/tester/
RUN chmod +x /bin/igo /home/tester/tester.sh
USER tester
WORKDIR /home/tester

ENV HOME=/home/tester

CMD ["-version"]