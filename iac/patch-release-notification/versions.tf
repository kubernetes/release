# Terraform version
terraform {
  backend "s3" {
    bucket = "tf-state-sig-release"
    key    = "cherry-pick-notification"
    region = "us-west-2"
  }

  required_version = "1.8.0"

  required_providers {
    ko = {
      source = "ko-build/ko"
    }

    aws = {
      source  = "hashicorp/aws"
      version = "5.51.1"
    }
  }
}
