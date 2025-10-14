FROM alpine AS build

RUN apk add --no-cache go npm openssl

WORKDIR /app

RUN openssl req -x509 -newkey rsa:2048 -keyout /tmp/key.pem -out /tmp/cert.pem -days 36500 -nodes -subj "/CN=example.com"

COPY package.json package-lock.json ./
RUN npm ci

COPY go.mod go.sum ./
RUN go mod download

COPY *.go /app/
ENV CGO_ENABLED=0
RUN go build -o /app/main

COPY ./views/ ./views/
COPY ./public/ ./public/

RUN npm run build

FROM alpine AS library

RUN apk add --no-cache npm ca-certificates

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --omit=dev

FROM scratch AS runtime

ARG user=1000
ARG group=1000

USER $user:$group
WORKDIR /app

COPY --from=library /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

COPY --from=build --chown=$user:$group /tmp/ /tmp/
COPY --from=build --chown=$user:$group /app/main .
COPY --from=library --chown=$user:$group /app/node_modules/ ./node_modules/

COPY --chown=$user:$group ./views/ ./views/
COPY --chown=$user:$group ./public/ ./public/
COPY --from=build --chown=$user:$group /app/public/style.output.css ./public/style.output.css

ENTRYPOINT ["/app/main"]
