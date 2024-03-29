VERSION 0.6
FROM golang:1.18-alpine
ARG NAME=earthly-secret-provider-vault
ARG DOCKER_REPO=ghcr.io/brumhard/$NAME
ARG BINPATH=/usr/local/bin/
ARG GOCACHE=/go-cache

local-setup:
    LOCALLY
    RUN git config --local core.hooksPath .githooks/

deps:
    WORKDIR /src
    ENV GO111MODULE=on
    ENV CGO_ENABLED=0
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

build:
    FROM +deps
    COPY --dir pkg/ cmd/ .
    ARG GOOS=linux
    ARG GOARCH=amd64
    ARG VARIANT
    RUN --mount=type=cache,target=$GOCACHE \
        GOARM=${VARIANT#"v"} go build -ldflags="-w -s" -o out/ ./...
    SAVE ARTIFACT out/*

build-local:
    ARG USEROS
    ARG USERARCH
    ARG USERVARIANT
    COPY --platform=linux/amd64 \
        (+build/$NAME --GOOS=$USEROS --GOARCH=$USERARCH --VARIANT=$USERVARIANT) /$NAME
    SAVE ARTIFACT /$NAME AS LOCAL out/$NAME

build-test:
    FROM +deps
    COPY --dir controllers/ pkg/ cmd/ .
    RUN --mount=type=cache,target=$GOCACHE \
        go build -ldflags="-w -s" -o /dev/null ./...

docker:
    ARG TARGETPLATFORM
    ARG TARGETOS
    ARG TARGETVARIANT
    ARG TARGETARCH
    FROM --platform=$TARGETPLATFORM \
        gcr.io/distroless/static:nonroot
    # use the following to not for multiarch with emulation as desribed in
    # https://docs.earthly.dev/docs/guides/multi-platform#creating-multi-platform-images-without-emulation
    COPY --platform=linux/amd64 \
        (+build/$NAME --GOOS=$TARGETOS --GOARCH=$TARGETARCH --VARIANT=$TARGETVARIANT) /usr/bin/$NAME
    USER 65532:65532
    # can't use variables in the entrypoint expression
    ENTRYPOINT ["/usr/bin/earthly-secret-provider-vault"]
    ARG EARTHLY_GIT_SHORT_HASH
    ARG DOCKER_TAG=$EARTHLY_GIT_SHORT_HASH
    SAVE IMAGE --push $DOCKER_REPO:$DOCKER_TAG

earthly:
    # earthly is only available as linux/amd64
    FROM earthly/earthly:v0.6.15
    COPY +build/$NAME /usr/bin/$NAME
    RUN earthly config global.secret_provider $NAME
    ARG EARTHLY_GIT_SHORT_HASH
    ARG DOCKER_TAG=$EARTHLY_GIT_SHORT_HASH
    SAVE IMAGE --push $DOCKER_REPO:$DOCKER_TAG-full

multiarch-docker:
    ARG EARTHLY_GIT_SHORT_HASH
    ARG DOCKER_TAG=$EARTHLY_GIT_SHORT_HASH
    BUILD --platform=linux/amd64 +docker --DOCKER_TAG=$DOCKER_TAG
    BUILD --platform=linux/arm64 +docker --DOCKER_TAG=$DOCKER_TAG
    BUILD --platform=linux/arm/v7 +docker --DOCKER_TAG=$DOCKER_TAG
    BUILD +earthly --DOCKER_TAG=$DOCKER_TAG

lint:
    ARG GOLANGCI_LINT_CACHE=/golangci-cache
    FROM +deps
    COPY +golangci-lint/golangci-lint $BINPATH
    COPY --dir pkg/ cmd/ .golangci.yml .
    RUN --mount=type=cache,target=$GOCACHE \
        --mount=type=cache,target=$GOLANGCI_LINT_CACHE \
        golangci-lint run -v ./...

test:
    FROM +deps
    COPY --dir pkg/ cmd/ .
    ARG GO_TEST="go test"
    RUN --mount=type=cache,target=$GOCACHE \
        eval "$GO_TEST ./..."

test-output:
    FROM +test --GO_TEST="go test -count 1 -coverprofile=cover.out"
    SAVE ARTIFACT cover.out

coverage:
    FROM +deps
    COPY --dir pkg/ cmd/ .
    COPY +test-output/cover.out .
    RUN go tool cover -func=cover.out

coverage-html:
    LOCALLY
    COPY +test-output/cover.out out/cover.out
    RUN go tool cover -html=out/cover.out

generate:
    COPY +build/docgen $BINPATH
    RUN mkdir -p docs && docgen -o docs
    SAVE ARTIFACT docs AS LOCAL docs

all:
    BUILD +deps
    BUILD +lint
    BUILD +coverage

trigger-release:
    ## TODO: don't rely on LOCALLY
    LOCALLY
    RUN git tag $(svu n)
    RUN git push --tags

release:
    FROM +deps
    RUN apk add --no-cache --update git
    COPY +svu/svu $BINPATH
    COPY +goreleaser/goreleaser $BINPATH
    COPY --dir .git/ pkg/ cmd/ .goreleaser.yaml .
    RUN svu current > VERSION
    RUN --push \
        --secret GITHUB_TOKEN \
        --mount=type=cache,target=$GOCACHE \
        goreleaser release --rm-dist --skip-validate
    BUILD --push +multiarch-docker --DOCKER_TAG=$(cat VERSION)


lint-commit:
    FROM node:alpine
    RUN apk --no-cache --update add git
    RUN npm install -g @commitlint/cli @commitlint/config-conventional
    WORKDIR /src
    COPY --dir .git/ .
    # check all commits to be in the right format
    RUN commitlint --to HEAD --verbose -x @commitlint/config-conventional

###########
# helper
###########

golangci-lint:
    FROM golangci/golangci-lint:v1.46.2
    SAVE ARTIFACT /usr/bin/golangci-lint

svu:
    FROM ghcr.io/caarlos0/svu:v1.9.0
    SAVE ARTIFACT /usr/local/bin/svu

goreleaser:
    FROM goreleaser/goreleaser:v1.9.2
    SAVE ARTIFACT /usr/bin/goreleaser
