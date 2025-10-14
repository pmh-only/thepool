FROM node:alpine AS assets
WORKDIR /app

COPY package.json package-lock.json ./

RUN --mount=type=cache,target=/root/.npm npm ci

COPY public/ ./public/
COPY views/ ./views/

RUN npm run build && npm prune --omit=dev

FROM scratch

ARG TARGETARCH
ARG user=1000
ARG group=1000

USER $user:$group
WORKDIR /app

COPY --chown=$user:$group app-linux-${TARGETARCH} /app/main
COPY --chown=$user:$group cert.pem key.pem /tmp/

COPY --from=assets --chown=$user:$group /app/public/ ./public/
COPY --from=assets --chown=$user:$group /app/node_modules/ ./node_modules/
COPY --chown=$user:$group views/ ./views/

ENTRYPOINT ["/app/main"]
