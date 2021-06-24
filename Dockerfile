FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.16-alpine3.13 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG MIN_COVERAGE_LEVEL=100

# Required to build the project
RUN apk update && \
    apk add \
        findutils \
        gcc \
        git \
        make \
        musl-dev

# Allows you to add additional packages via build-arg
ARG ADDITIONAL_PACKAGE
ARG GOPROXY=""
ARG GOFLAGS=""

ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ENV GO111MODULE="on"
ENV CGO_ENABLED=1

WORKDIR /src
COPY . .

RUN make bins
RUN make test | tee test_report.log
RUN grep 'coverage:' test_report.log | \
    sed -r 's/.*?coverage: ([0-9]+\.[0-9]+%) .*/\1/' | \
    awk -f coverage_check.awk -v min_cov_level=${MIN_COVERAGE_LEVEL}
RUN make fmtcheck && \
    make staticcheck && \
    make errcheck

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.13
RUN apk --no-cache add ca-certificates && \
    addgroup -S app && \
    adduser -S -g app app && \
    mkdir -p /home/app && \
    chown app /home/app

WORKDIR /home/app

COPY --from=builder /src/bin/cmd openfaas-sqs-connector
RUN chown -R app /home/app

USER app

ENTRYPOINT ["./openfaas-sqs-connector"]
CMD ["--help"]
