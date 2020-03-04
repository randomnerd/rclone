FROM golang AS builder

COPY . /go/src/github.com/rclone/rclone/
WORKDIR /go/src/github.com/rclone/rclone/

#RUN make quicktest
RUN \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  make
RUN ./rclone version

# Begin final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates fuse
ENV BUCKET=""
ENV GUI_USER=storj
ENV GUI_PASSWORD=changeme
ENV GUI_LISTEN=0.0.0.0:5572
ENV GUI_ALLOW_ORIGIN="*"

COPY --from=builder /go/src/github.com/rclone/rclone/rclone /usr/local/bin/
COPY ./contrib/docker/entrypoint.sh /entrypoint.sh

WORKDIR /data
ENV XDG_CONFIG_HOME=/config
EXPOSE 5572

ENTRYPOINT [ "ash", "/entrypoint.sh" ]
