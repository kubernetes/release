# Image Vulnerability Dashboard

The vulnerability dashboard is intended to surface the vulnerabilities of
images in the Kubernetes production project - k8s-artifacts-prod. The dashboard
operates as a static HTML page hosted on Google Cloud Storage.

It can be accessed by visiting this
[page](https://storage.googleapis.com/k8s-artifacts-prod-vuln-dashboard/dashboard.html).

## Information Flow

### CAS Adapter

The dashboard utilizes the Container Analysis Service in order to actually get
the vulnerabilities it displays. However, the vulnerability information can't
be used as provided by CAS. This is because CAS provides lots of unnecessary
information which make it difficult to parse into an HTML table; not to
mention, the large file size due to the sheer amount of information for each of
Kubernetes' production images.

In order to use this information, we use an adapter which processes the
vulnerability occurrences returned from the CAS into a new
[struct](/pkg/vulndash/adapter/types.go) containing only the info that the
dashboard needs in order to create its table. This information is then uploaded
as a JSON file to Google Cloud Storage, where it can be parsed into an
easy-to-read HTML table. The adapter is implemented in
[`/pkg/vulndash/adapter`](/pkg/vulndash/adapter/adapter.go).

### JS Parser

The CAS adapter described above writes the processed vulnerability information
to a JSON stored in the vulnerability dashbaord's GCS bucket. In order to
convert this JSON to a HTML table, a simple JavaScript file is also placed in
the GCS bucket which can read in the contents of the JSON and create the table.

The most updated versions of both the [JavaScript](dashboard.js) file and the
static [HTML](dashboard.html) page are also uploaded to Google Cloud Storage
whenever the adapter runs.

## Integration With Prow

In order to have the dashboard display the most up to date vulnerability
information from the Container Analysis Service, a
[periodic](https://git.k8s.io/test-infra/prow/jobs.md) Prow job has been set up
([ci-k8sio-vuln-dashboard-update](https://git.k8s.io/test-infra/config/jobs/kubernetes/wg-k8s-infra/trusted/wg-k8s-infra-trusted.yaml)).

This Prow job runs once every 24 hours, and runs the adapter in order to get
the most recent vulnerabilities and upload any new updates to the dashboard
files stored in Google Cloud Storage.
