Table of Contents
=================
* [Intro](#intro)
* [Typical Workflows](#typical-workflows)

# Container Builder Staging and Release

## Intro

The Kubernetes release process can now be staged and released on both the GCP
Container Builder (GCB) or on the desktop with the necessary permissions.

## Typical Workflows

The typical workflow is very simple and works similar to `anago` in both mock
and `--nomock` variants with a clear division between the two.  Stage and
release using either `--nomock` or in the default (mock) mode.

The hybrid model is also supported
* Stage on GCB
* Release on desktop

```
# On GCB, stage a (mock) master branch build from head
$ gcbmgr stage master --build-at-head

# On GCB, stage a (mock) release-1.9 branch build using test signal
$ gcbmgr stage release-1.9

# On GCB, view last 5 jobs
$ gcbmgr
-OR-
$ gcbmgr list

# View completed staged builds
$ gcbmgr staged

# Release (from GCB)
$ gcbmgr release master --buildversion=<a staged master build version>

# Release (from GCB)
$ gcbmgr release release-1.9 --buildversion=<a staged release-1.9 build version>

# And of course the man page has all the most detailed and up to date info:
$ gcbmgr -man
```

Guidance from `gcbmgr staged` instructs you how to release a staged build on
GCB or the desktop.

**NOTE:**
Releases from GCB are currently unable to send email, so the update
occurs in the form of a new release tracking issue on the
kubernetes/sig-release repo (k8s-release-robot/sig-release for mock runs).

To send the standard e-mail announcement after a release on GCB, you may
use the helper `release-notify` which will grab the notification details
from gs://<BUCKET>/archive/anago-$VERSION/announcement.html:
```
$ release-notify $VERSION
```

This works like the rest of the tool and honors --nomock and prompts
before any mail is sent.
