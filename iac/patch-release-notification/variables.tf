variable "region" {
  description = "The AWS region to deploy the resources"
  type        = string
}

variable "email_identity" {
  description = "The email address or domain to verify"
  type        = string
}

variable "to_email" {
  description = "The email address to send the notification"
  type        = string
}

variable "no_mock" {
  description = "if will send the message to dev@kubernetes.io or just internal"
  type        = bool
}

variable "days_to_alert" {
  description = "when to send the notification"
  type        = number
}

variable "schedule_path" {
  description = "path where we can find the schedule.yaml"
  type        = string
}

variable "repository" {
  description = "The ECR repository to use for the image"
  type        = string
  default     = ""
}
