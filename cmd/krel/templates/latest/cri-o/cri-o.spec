%global debug_package %{nil}
%undefine _missing_build_ids_terminate_build

Name: cri-o
Version: {{ .RPMVersion }}
Release: {{ .Revision }}
Summary: Open Container Initiative-based implementation of Kubernetes Container Runtime Interface

%if "%{_vendor}" == "debbuild"
Group: admin
%endif

Packager: Kubernetes Authors <dev@kubernetes.io>
License: Apache-2.0
URL: https://kubernetes.io
Source0: %{name}_%{version}.orig.tar.gz
BuildRequires: sed

%description
%{summary}.

%prep
%setup -q -c

%build
# Nothing to build

%install
%define archive_root "$(uname -m)"/cri-o

# Binaries
install -dp %{buildroot}%{_bindir}
install -p -m 755 %{archive_root}/bin/crio %{buildroot}%{_bindir}/crio
install -p -m 755 %{archive_root}/bin/conmon %{buildroot}%{_bindir}/conmon
install -p -m 755 %{archive_root}/bin/pinns %{buildroot}%{_bindir}/pinns
install -p -m 755 %{archive_root}/bin/crun %{buildroot}%{_bindir}/crun

# Completions
install -d -m 755 %{buildroot}%{_datadir}/bash-completion/completions
install -D -m 644 -t %{buildroot}%{_datadir}/bash-completion/completions %{archive_root}/completions/bash/crio

install -d -m 755 %{buildroot}%{_datadir}/fish/completions
install -D -m 644 -t %{buildroot}%{_datadir}/fish/completions %{archive_root}/completions/fish/crio.fish

install -d -m 755 %{buildroot}%{_datadir}/zsh/site-functions
install -D -m 644 -t %{buildroot}%{_datadir}/zsh/site-functions %{archive_root}/completions/zsh/_crio

# Configurations
install -dp %{buildroot}%{_sysconfdir}/containers
install -p -m 644 %{archive_root}/contrib/policy.json %{buildroot}%{_sysconfdir}/containers/policy.json

install -p -m 644 %{archive_root}/etc/crictl.yaml %{buildroot}%{_sysconfdir}/crictl.yaml

install -dp %{buildroot}%{_sysconfdir}/crio/crio.conf.d
install -p -m 644 %{archive_root}/etc/10-crun.conf %{buildroot}%{_sysconfdir}/crio/crio.conf.d/10-crun.conf
install -p -m 644 %{archive_root}/etc/crio.conf %{buildroot}%{_sysconfdir}/crio/crio.conf

install -dp %{buildroot}%{_datadir}/oci-umount/oci-umount.d
install -p -m 644 %{archive_root}/etc/crio-umount.conf %{buildroot}%{_datadir}/oci-umount/oci-umount.d/crio-umount.conf

install -dp %{buildroot}%{_sysconfdir}/cni/net.d
install -p -m 644 %{archive_root}/contrib/11-crio-ipv4-bridge.conflist %{buildroot}%{_sysconfdir}/cni/net.d/11-crio-ipv4-bridge.conflist

# Fix the prefix in crio.service
sed -i 's;/usr/local/bin;/usr/bin;g' %{archive_root}/contrib/crio.service
install -D -m 644 -t %{buildroot}%{_unitdir} %{archive_root}/contrib/crio.service

# Docs
install -D -m 644 -t %{buildroot}%{_mandir}/man5 %{archive_root}/man/crio.conf.5
install -D -m 644 -t %{buildroot}%{_mandir}/man5 %{archive_root}/man/crio.conf.d.5
install -D -m 644 -t %{buildroot}%{_mandir}/man8 %{archive_root}/man/crio.8

%files
# Binaries
%{_bindir}/crio
%{_bindir}/conmon
%{_bindir}/pinns
%{_bindir}/crun

# Completions
%{_datadir}/bash-completion/completions/crio
%dir %{_datadir}/fish
%dir %{_datadir}/fish/completions
%{_datadir}/fish/completions/crio.fish
%dir %{_datadir}/zsh
%dir %{_datadir}/zsh/site-functions
%{_datadir}/zsh/site-functions/_crio

# Configurations
%dir %{_sysconfdir}/containers
%config(noreplace) %{_sysconfdir}/containers/policy.json
%config(noreplace) %{_sysconfdir}/crictl.yaml
%dir %{_sysconfdir}/cni
%dir %{_sysconfdir}/cni/net.d
%config(noreplace) %{_sysconfdir}/cni/net.d/11-crio-ipv4-bridge.conflist
%{_unitdir}/crio.service
%dir %{_sysconfdir}/crio
%dir %{_sysconfdir}/crio/crio.conf.d
%{_sysconfdir}/crio/crio.conf
%{_sysconfdir}/crio/crio.conf.d/10-crun.conf
%dir %{_datadir}/oci-umount
%dir %{_datadir}/oci-umount/oci-umount.d
%{_datadir}/oci-umount/oci-umount.d/crio-umount.conf

# Docs
%{_mandir}/man5/crio.conf*5*
%{_mandir}/man8/crio*.8*

%license %{_arch}/cri-o/LICENSE
%doc %{_arch}/cri-o/README.md

%changelog
