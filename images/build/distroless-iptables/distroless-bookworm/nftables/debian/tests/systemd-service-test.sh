#!/bin/sh

set -ex

SERVICE=nftables.service

# The testsuite requires kernel at least 5.x
if [ "$(uname -r | cut -d. -f1)" -lt 5 ] ; then
	: WARNING this testsuite is likely to produce many fails because of old kernel, ending now
	exit 0
fi

systemctl_call()
{
	if systemctl $1 $SERVICE ; then
		return 0
	else
		journalctl -u $SERVICE
		return 1
	fi
}

# package ships service disabled by default
if ! systemctl_call enable ; then
	: WARNING enabling the service failed
fi

if systemctl -q is-active $SERVICE ; then
	: WARNING initial service running, stopping now
	if ! systemctl_call stop ; then
		: ERROR unable to stop the initial service
		exit 1
	fi
fi

if [ $(nft list ruleset | wc -l) -ne 0 ] ; then
	: WARNING initial ruleset is not empty, flushing now
	nft flush ruleset
fi

if ! systemctl_call start ; then
	: ERROR failed to start systemd service
	exit 1
fi
if [ $(nft list ruleset | wc -l) -eq 0 ] ; then
	: ERROR no ruleset loaded after systemd service start
	exit 1
fi

systemctl_call status
nft list ruleset

if ! systemctl_call stop ; then
	: ERROR failed to stop systemd service
	exit 1
fi
if [ $(nft list ruleset | wc -l) -ne 0 ] ; then
	: ERROR ruleset still loaded after systemd service stop
	exit 1
fi

if ! systemctl_call restart ; then
	: ERROR failed to restart systemd service
	exit 1
fi
if [ $(nft list ruleset | wc -l) -eq 0 ] ; then
	: ERROR no ruleset loaded after systemd service restart
	exit 1
fi

: INFO test was OK
exit 0
