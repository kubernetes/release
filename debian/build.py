#!/usr/bin/env python3

import logging
import sys
import subprocess

# to build multiarch you might need to add the architectures
# $ sudo dpkg --add-architecture armhf
# $ sudo dpkg --add-architecture arm64

FORMAT = '+ %(asctime)-15s %(message)s'
logging.basicConfig(format=FORMAT)
LOG = logging.getLogger()
LOG.setLevel(0)

class Build():

  def __init__(self, pkg, arch, distro, version, revision):
    self.arch = arch
    self.distro = distro
    self.pkg = pkg
    self.version = version
    self.revision = revision

  def run(self):
    cmd = (
      'go',
      'run',
      'build.go',
      '-arch', self.arch,
      '-distro_name', self.distro,
      '-package', self.pkg,
      '-version', self.version,
      '-revision', self.revision
    )
    LOG.debug("running cmd: %s", cmd)
    return subprocess.call(cmd)

  def __str__(self):
    return "%s(arch=%s,distro=%s,version=%s,revision=%s)" % (
        self.pkg, self.arch, self.distro, self.version, self.revision)

def main():
  architectures=('amd64', 'arm', 'arm64')
  distros=('xenial',)

  packages = {
    'kubectl': (
      ('1.3.7', '00'),
      ('1.4.0-beta.5', '00'),
    ),
    'kubelet': (
      ('1.3.7', '00'),
      ('1.4.0-beta.5', '00'),
    ),
    'kubeadm': (
      ('1.4.0-alpha', '00'),
    ),
    'kubernetes-cni': (
      ('0.3.0.1-07a8a2', '00'),
    ),
  }

  builds = []

  for pkg, versions in packages.items():
    for version, revision in versions:
      for arch in architectures:
        for distro in distros:
          builds.append(Build(pkg, arch, distro, version, revision))

  for build in builds:
    LOG.debug("planning to build: %s" % build)

  for build in builds:
    LOG.info("building: %s", build)
    rc = build.run()
    if rc:
      LOG.error("error building: %s" % build)
      sys.exit(rc)
    LOG.debug("successfuly built: %s" % build)


if __name__ == '__main__':
  main()
