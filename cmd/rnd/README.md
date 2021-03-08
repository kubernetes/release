# Release Notes Daemon (rnd)

This is an experimental project, reimagining release notes review and revision through version control, adopting the GitOps philosophy.

Details are available in [`kubernetes/release#1889`](https://github.com/kubernetes/release/issues/1889). This project is incomplete at the moment.

Project status:
- [x] Implement `fetch` and save content as YAML
- [ ] Implement `render` where given a Git repository path and release version, render the release notes document
