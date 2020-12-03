# publish-release

## A tool to publish releases

`publish-release` is a command to publish and announce software releases. 
The current implementation implements a subcommand to update a release page
un GitHub, future plans include incorporating SIG Release's announce email 
tool in a generic form.

### Update a GitHub release page

This command allows you to update a release page on GitHub. You can define 
your logo, name of the release, a link to your changelog and an introductory
text.

You can customize the look of the release page look by using a custom golang
template with any number of string substitutions.

```
This command updates the GitHub release page for a given tag. It will
update the page using a built in template or you can update it using
a custom template.

Before updating the page, the tag has to exist already on github.

To publish the page, --nomock has to be defined. Otherwise, the rendered
page will be printed to stdout and the program will exit.

CUSTOM TEMPLATES
================
You can define a custom golang template to use in your release page. Your
template can contain string substitutions and you can define those using 
the --substitution flag:

  --substitution="releaseTheme:Accentuate the Paw-sitive"
  --substitution="releaseLogo:accentuate-the-pawsitive.png"

ASSET FILES
===========
This command supports uploading release assets to the github page. You
can add asset files with the --asset flag:

  --asset=_output/kubernetes-1.18.2-2.fc33.x86_64.rpm

You can also specify a label for the assets by appending it with a colon
to the asset file:

  --asset="_output/kubernetes-1.18.2-2.fc33.x86_64.rpm:RPM Package for amd64"

Usage:
  release-announce github [flags]

Flags:
  -a, --asset strings          Path to asset file for the release. Can be specified multiple times.
      --draft                  Mark the release as a draft in GitHub so you can finish editing and publish it manually.
  -h, --help                   help for github
  -n, --name string            name for the release
      --noupdate               Fail if the release already exists
  -r, --repo string            repository slug containing the release page
  -s, --substitution strings   String substitution for the page template
      --template string        path to a custom page template

Global Flags:
      --log-level string   the logging verbosity, either 'panic', 'fatal', 'error', 'warn', 'warning', 'info', 'debug' or 'trace' (default "info")
      --nomock             tag for the release to be used
  -t, --tag string         tag for the release to be used

```