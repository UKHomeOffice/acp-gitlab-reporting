FROM golang:1.20 as builder

WORKDIR /go/src/github.com/UKHomeOffice/acp-gitlab-reporting

COPY go.mod ./

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go install -v \
            github.com/UKHomeOffice/acp-gitlab-reporting

FROM alpine:3.17
RUN apk --no-cache add ca-certificates

RUN addgroup -g 1000 -S app && \
    adduser -u 1000 -S app -G app

USER 1000

COPY --from=builder /go/bin/acp-gitlab-reporting /acp-gitlab-reporter
CMD ["/acp-gitlab-reporter"]
