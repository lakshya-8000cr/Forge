terraform {
  backend "s3" {
    bucket = "forge-tfstate-lakshya-v2-xyz-892"
    key    = "forge/dev/terraform.tfstate"
    region = "eu-north-1"

    use_lockfile = true

    encrypt = true
  }
}