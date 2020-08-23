---
title: Deprecation warnings
sig: api-machinery
---
SIG API Machinery implemented warning mechanisms when using deprecated APIs that are visible to API consumers and metrics visible to cluster administrators. Requests to a deprecated API are returned with a warning containing a target removal release and any replacement API.
