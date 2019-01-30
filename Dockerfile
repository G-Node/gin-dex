FROM golang:alpine
RUN apk --no-cache --no-progress add build-base git
RUN apk --no-cache --no-progress add curl

# repositories from GOGS for indexing
# should be bound read-only
VOLUME ["/repos"]

RUN go version


COPY ./go.mod ./go.sum /gindex/
WORKDIR /gindex
# download deps before bringing in the main package
RUN go mod download

COPY ./cmd /gindex/cmd/
RUN go build ./cmd/gindex

ENTRYPOINT ./gindex
EXPOSE 10443
