FROM golang:1.26-alpine AS build

ARG VERSION=0.1
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/reelsieve ./cmd/reelsieve && mkdir -p /out/data && chown 65532:65532 /out/data

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/reelsieve /reelsieve
COPY --from=build --chown=65532:65532 /out/data /data
USER 65532:65532
WORKDIR /data
EXPOSE 8080
ENTRYPOINT ["/reelsieve"]
