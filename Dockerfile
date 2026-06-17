FROM golang:1.26-alpine AS build

ARG VERSION=0.1
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/reelsieve ./cmd/reelsieve

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/reelsieve /reelsieve
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/reelsieve"]
