Name: cri-tools
Version: {{ .Version }}
Release: {{ .Revision }}
Summary: Command-line utility for interacting with a container runtime.

License: ASL 2.0
URL: https://kubernetes.io
Source0: https://storage.googleapis.com/k8s-artifacts-cri-tools/release/v{{ .Version }}/crictl-v{{ .Version }}-linux-{{ .GoArch }}.tar.gz

BuildRequires: systemd
BuildRequires: curl

%description
Command-line utility for interacting with a container runtime.

%prep
%setup -c -a 7 -T -n cri-tools

# TODO: Do we need these?
#%autosetup
#%build
#%configure
#%make_build

%install
# TODO: Do we need this?
#rm -rf $RPM_BUILD_ROOT

cd %{_builddir}
install -p -m 755 -t %{buildroot}%{_bindir}/ cri-tools/crictl

# TODO: Do we need this?
#%make_install

%files
%{_bindir}/crictl

# TODO: Do we need these?
#%license add-license-file-here
#%doc add-docs-here


%changelog
* Sat Jan  4 2020 Stephen Augustus <saugustus@vmware.com> - 1.18.0
- Create separate spec file for cri-tools
