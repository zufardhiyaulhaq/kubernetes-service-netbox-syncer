#################
FROM alpine:3.22.0 as kubernetes-service-netbox-syncer-base

USER root

RUN addgroup -g 10001 kubernetes-service-netbox-syncer && \
    adduser --disabled-password --system --gecos "" --home "/home/kubernetes-service-netbox-syncer" --shell "/sbin/nologin" --uid 10001 kubernetes-service-netbox-syncer && \
    mkdir -p "/home/kubernetes-service-netbox-syncer" && \
    chown kubernetes-service-netbox-syncer:0 /home/kubernetes-service-netbox-syncer && \
    chmod g=u /home/kubernetes-service-netbox-syncer && \
    chmod g=u /etc/passwd
RUN apk add --update --no-cache alpine-sdk curl

ENV USER=kubernetes-service-netbox-syncer
USER 10001
WORKDIR /home/kubernetes-service-netbox-syncer

#################
# Builder image
#################
FROM golang:1.24-bullseye AS kubernetes-service-netbox-syncer-builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build main.go

#################
# Final image
#################
FROM kubernetes-service-netbox-syncer-base

COPY --from=kubernetes-service-netbox-syncer-builder /app/main /usr/local/bin/kubernetes-service-netbox-syncer

# Command to run the executable
ENTRYPOINT ["kubernetes-service-netbox-syncer"]