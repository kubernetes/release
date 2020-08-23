---
title: Avoiding permanent beta
sig: Architecture
---
From Kubernetes 1.20 onwards, SIG Architecture will implement a new policy to transition all REST APIs out of beta within nine months. The idea behind the new policy is to avoid features staying in beta for a long time. Once a new API enters beta, it will have nine months to either:

  - reach GA, and deprecate the beta, or
  - have a new beta version (and deprecate the previous beta).

If a REST API reaches the end of that nine-month countdown, then the next Kubernetes release will deprecate that API version. More information can be found on the [Kubernetes Blog][blog-post].

[blog-post]: https://deploy-preview-21274--kubernetes-io-master-staging.netlify.app/blog/2020/08/21/moving-forward-from-beta/