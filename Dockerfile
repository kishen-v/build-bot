FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:1.27 AS builder

ARG TARGETOS
ARG TARGETARCH
ARG ARCH

WORKDIR /workspace
COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-${ARCH}} go build -ldflags '-s -w' -trimpath -o /build-bot .

FROM gcr.io/distroless/static
WORKDIR /
COPY --from=builder /build-bot /build-bot
ENTRYPOINT ["/build-bot"]
