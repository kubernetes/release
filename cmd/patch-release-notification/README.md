# Patch Release Notification

AWS Lambda function that emails the community when the cherry-pick deadline for
the next patch releases is approaching.

It runs once a day, reads the patch release
[`schedule.yaml`](https://github.com/kubernetes/website/blob/main/data/releases/schedule.yaml),
and sends one email through Amazon SES listing every release branch whose next
cherry-pick deadline is exactly `DAYS_TO_ALERT` calendar days (UTC) away. If no
branch matches, nothing is sent.

The infrastructure (Lambda, SES identity, IAM role and EventBridge schedule) is
managed in
[kubernetes/k8s.io](https://github.com/kubernetes/k8s.io/tree/main/infra/aws/terraform/sig-release).

## Configuration

| Variable        | Required | Description                                                                                                 |
|-----------------|----------|-------------------------------------------------------------------------------------------------------------|
| `FROM_EMAIL`    | yes      | SES verified identity used as sender.                                                                       |
| `TO_EMAIL`      | yes      | Recipient, for example `dev@kubernetes.io`.                                                                 |
| `SCHEDULE_PATH` | yes      | `https://` URL or local path to `schedule.yaml`. Plain `http://` is rejected.                               |
| `DAYS_TO_ALERT` | yes      | Number of calendar days before the cherry-pick deadline to send the email.                                  |
| `NO_MOCK`       | no       | When `true`, the email goes to `TO_EMAIL`. Defaults to `false`, which sends it to `FROM_EMAIL` only so it can be checked first. |

The AWS region comes from the `AWS_REGION` variable that the Lambda runtime sets.
The function needs the `ses:SendEmail` permission on the `FROM_EMAIL` identity.

The function ignores the event payload, so any scheduled trigger works.
