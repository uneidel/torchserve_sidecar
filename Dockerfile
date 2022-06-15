FROM golang:alpine AS build-env

ENV GO111MODULE=on

RUN apk --no-cache add  build-base git  mercurial gcc
ADD ./src/ /src
RUN cd /src && go build -o torchserveinitcontainer

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/torchserveinitcontainer /app/


CMD ["/app/torchserveinitcontainer"]