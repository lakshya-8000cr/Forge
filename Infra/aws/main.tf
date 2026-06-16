terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  region = "eu-north-1"
}

resource "aws_s3_bucket" "forge_test" {
  bucket = "forge-test-lakshya-v2-xyz-892"
}

resource "aws_s3_bucket" "tf_state" {
  bucket = "forge-tfstate-lakshya-v2-xyz-892"
}



