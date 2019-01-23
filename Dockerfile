FROM golang:alpine
RUN apk --no-cache --no-progress add --virtual build-deps build-base git

# repositories from GOGS for indexing
# should be bound read-only
VOLUME ["/repos"]

COPY . /gin-dex
WORKDIR /gin-dex
RUN go build .

ENTRYPOINT ./gin-dex
EXPOSE 10443
