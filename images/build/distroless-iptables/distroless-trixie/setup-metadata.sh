#!/bin/bash

# Copyright 2026 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

umask 0022

# Create the copyright file
mkdir -p /usr/share/doc/nftables
cat > /usr/share/doc/nftables/copyright <<EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: nftables
Source: http://git.netfilter.org/nftables
Comment:
 Simplified version of Debian's /usr/share/doc/nftables/copyright

Files: *
Copyright: 1999-2026 NFTables Authors
License: GPL-2

License: GPL-2
 This program is free software; you can redistribute it and/or modify
 it under the terms of the GNU Library General Public License as published by
 the Free Software Foundation.
 .
 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Library General Public License for more details.
 .
 You should have received a copy of the GNU General Public License
 along with this program. If not, see <http://www.gnu.org/licenses/>
 .
 On Debian systems, the complete text of the GNU General
 Public License version 2 can be found in "/usr/share/common-licenses/GPL-2".
EOF

# Create the file list(s)
cat > /var/lib/dpkg/info/nftables.list <<EOF
/usr/sbin/nft
/usr/share/doc/nftables/copyright
EOF
(cd /; md5sum usr/sbin/nft usr/share/doc/nftables/copyright > /var/lib/dpkg/info/nftables.md5sums)

# Set the remaining dpkg metadata
cat >> /var/lib/dpkg/status <<EOF
Package: nftables
Status: install ok installed
Maintainer: Kubernetes Release Team <release-team@kubernetes.io>
Architecture: $(dpkg --print-architecture)
Version: 1.0.6.1
Description: Locally-built version of upstream stable nftables release
EOF
