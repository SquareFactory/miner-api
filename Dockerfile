# ------------------------------------
FROM docker.io/library/golang:1.20-alpine as api-builder
# ------------------------------------

WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN apk add --no-cache make
RUN make build-all

# ------------------------------------
FROM ghcr.io/squarefactory/slurm:latest-login-rocky8.6 as slurm-login
# ------------------------------------

COPY --from=api-builder /work/bin/miner-api /usr/sbin/miner-api
COPY slurm/s6-rc.d/miner-api/ /etc/s6-overlay/s6-rc.d/miner-api/

RUN rm -rf /etc/s6-overlay/s6-rc.d/ssh/ \
    && rm -rf /etc/s6-overlay/s6-rc.d/user/contents.d/ssh \
    && touch /etc/s6-overlay/s6-rc.d/user/contents.d/miner-api

ENV S6_KEEP_ENV=1

ENTRYPOINT ["/init"]
