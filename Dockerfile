FROM golang AS builder

WORKDIR /go/src/github.com/rclone/rclone/
COPY backend ./backend
COPY bin ./bin
COPY cmd ./cmd
COPY fs ./fs
COPY lib ./lib
COPY vfs ./vfs
COPY vendor ./vendor
COPY rclone.go go.* Makefile VERSION ./
COPY .git ./.git
RUN ls -l /go/src/github.com/rclone/rclone/

#RUN make quicktest 2>/dev/null && echo TESTS OK
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make rclone
RUN ./rclone version

# Begin final image
FROM alpine:latest

ENV STORJ_SCOPE=$STORJ_SCOPE
RUN apk --no-cache add ca-certificates fuse bash

COPY --from=builder /go/bin/rclone /usr/local/bin/
COPY ./contrib/docker/entrypoint.sh /entrypoint.sh

WORKDIR /data
ENV XDG_CONFIG_HOME=/config
EXPOSE 5572
ENTRYPOINT [ "/entrypoint.sh" ]

