#!/bin/ash
ls -la / && ls -la /data
set -u
export RCLONE="/usr/local/bin/rclone --config=/config/rclone/rclone.conf"
mkdir -p /{config,cache,data}
mkdir -p /config/rclone
#mkdir -p /cache
#mkdir -p /data

# Make sure we have scope token defined
test -z ${STORJ_SCOPE} &&\
	echo 'Please set $STORJ_SCOPE environment variable to your Storj sharing token (run "uplink setup && uplink share" to get one)' &&\
	exit 1

# Make sure we have config file created
test -x /config/rclone.conf ||\
	${RCLONE} config create storj storj scope ${STORJ_SCOPE} defaults release skip-peer-ca-whitelist false

# Mount our endpoint
${RCLONE} mount storj://${BUCKET} /data \
	--allow-non-empty \
	--allow-other \
	--write-back-cache \
	--vfs-case-insensitive \
	--vfs-cache-mode=full \
	--attr-timeout=10s \
	--rc \
	--rc-no-auth \
	--rc-addr=${GUI_LISTEN} \
	--rc-allow-origin="${GUI_ALLOW_ORIGIN}" \
	--rc-user=${GUI_USER} \
	--rc-pass=${GUI_PASSWORD} \
	--rc-serve \
	--rc-web-gui \
	--rc-web-gui-update \
	--rc-web-gui-no-open-browser \
	--track-renames \
	--update \
#	--use-mmap \
#	--use-cookies \
	${@}

