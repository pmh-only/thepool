FROM scratch

ARG TARGETARCH
ARG user=1000
ARG group=1000

USER $user:$group
WORKDIR /app

COPY --chown=$user:$group app-linux-${TARGETARCH} /app/main

COPY --chown=$user:$group tmp/ /tmp/
COPY --chown=$user:$group views/ ./views/
COPY --chown=$user:$group public/ ./public/
COPY --chown=$user:$group node_modules/ ./node_modules/

ENTRYPOINT ["/app/main"]
