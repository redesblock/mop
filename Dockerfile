FROM golang:1.17 AS build

WORKDIR /src

COPY . ./

RUN make

FROM debian:11.2-slim

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*; \
    groupadd -r hop --gid 999; \
    useradd -r -g hop --uid 999 --no-log-init -m hop;

# make sure mounted volumes have correct permissions
RUN mkdir -p /home/hop/.hop && chown 999:999 /home/hop/.hop

COPY --from=build /src/bin/hop /usr/local/bin/hop

EXPOSE 1633 1634 1635
USER hop
WORKDIR /home/hop
VOLUME /home/hop/.hop

ENTRYPOINT ["hop"]
