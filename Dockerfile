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
    groupadd -r mop --gid 999; \
    useradd -r -g mop --uid 999 --no-log-init -m mop;

# make sure mounted volumes have correct permissions
RUN mkdir -p /home/mop/.mop && chown 999:999 /home/mop/.mop

COPY --from=build /src/bin/mop /usr/local/bin/mop

EXPOSE 1633 1634 1635
USER mop
WORKDIR /home/mop
VOLUME /home/mop/.mop

ENTRYPOINT ["mop"]
