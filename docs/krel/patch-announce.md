# krel patch-anounce
Send out patch release announcement emails

- [Summary](#summary)
- [Installation](#installation)
- [Usage](#usage)
- [Important notes](#important-notes)


## Summary
The `krel patch-announce` subcommand sends out email messages to notify developers about a release. To send the email messages, it needs a Sendgrid API Key and an GitHub token

## Installation
Tu run the `patch-announce` subcommand, [install krel](README.md#installation). You will need a Sendgrid token to be able to send messages and a GitHub token to generate the release notes.

## Usage
```
  krel patch-announce [flags]
```

### Command line flags
```
  -c, --cut-date string           date when the patch release is planned to be cut
  -f, --freeze-date string        date when no CPs are allowed anymore
  -g, --github-token string       a GitHub token, used r/o for generating the release notes (default $GITHUB_TOKEN)
  -h, --help                      help for patch-announce
  -r, --release-repo string       local path of the k/release checkout (default "./release")
  -e, --sender-email string       email sender's address
  -n, --sender-name string        email sender's name
  -s, --sendgrid-api-key string   API key for sendgrid
```

### Notification recipients

Currently, `patch-announce` will notify the following addresses:
| Recipient | E-Mail Address |
| --------- | -------------- |
| Kubernetes developer/contributor discussion | kubernetes-dev@googlegroups.com |
| kubernetes-dev-announce | kubernetes-dev-announce@googlegroups.com |


