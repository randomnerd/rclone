#!/bin/bash
set -u
RCLONE="/usr/local/bin/rclone --config=/config/rclone.conf --cache-dir=/cache"
function ensureConfig
{
	local scope="${STORJ_SCOPE}"
	if [[ -z "${scope}" ]] && [[ ! -x /config/rclone.conf ]]
	then
		echo "Please set \$STORJ_SCOPE environment variable to your Storj sharing token (run 'uplink setup && uplink share' to get one)" && exit 1
	elif [[ ! -x "/config/rclone.conf" ]]
		/bin/sh -c "${RCLONE} config create storj storj scope ${STORJ_SCOPE} skip-peer-ca-whitelist true defaults dev"
	
	fi;

}

ensureConfig
#${RCLONE} mount storj:/ /data --write-back-cache --allow-other --no-modtime --no-checksum --vfs-case-insensitive --vfs-cache-mode=full --attr-timeout=10s --fast-list --no-check-certificate --no-check-dest --no-update-modtime --rc --rc-addr=0.0.0.0:5572 --rc-serve --rc-web-gui --rc-web-gui-force-update --rc-web-gui-no-open-browser --track-renames --update --use-mmap --use-cookies -vv --rc-allow-origin="*" --rc-no-auth
${RCLONE} mount storj://storj /data --write-back-cache --no-modtime --no-checksum --vfs-case-insensitive --vfs-cache-mode=full --attr-timeout=10s --fast-list --no-check-certificate --no-check-dest --no-update-modtime --rc-user=${GUI_USER:-"storj"} --rc --rc-addr=${GUI_LISTEN:-"127.0.0.1:5572"} --rc-serve --rc-web-gui --rc-web-gui-update --rc-web-gui-no-open-browser --track-renames --update --use-mmap --use-cookies --rc-allow-origin="${GUI_ALLOW_ORIGIN:-"*"}" --rc-pass=${GUI_PASSWORD:-"changeme"} $@

