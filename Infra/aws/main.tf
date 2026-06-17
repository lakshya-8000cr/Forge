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

resource "aws_vpc" "forge_vpc" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "forge-vpc"
  }
}

resource "aws_subnet" "forge_public_subnet" {
  vpc_id                  = aws_vpc.forge_vpc.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "eu-north-1a"
  map_public_ip_on_launch = true

  tags = {
    Name = "forge-public-subnet-1"
  }
}

resource "aws_subnet" "forge_public_subnet_2" {
  vpc_id                  = aws_vpc.forge_vpc.id
  cidr_block              = "10.0.2.0/24"
  availability_zone       = "eu-north-1b"
  map_public_ip_on_launch = true

  tags = {
    Name = "forge-public-subnet-2"
  }
}

resource "aws_internet_gateway" "forge_igw" {
  vpc_id = aws_vpc.forge_vpc.id

  tags = {
    Name = "forge-igw"
  }
}

resource "aws_route_table" "forge_public_rt" {
  vpc_id = aws_vpc.forge_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.forge_igw.id
  }

  tags = {
    Name = "forge-public-route-table"
  }
}

resource "aws_route_table_association" "forge_public_assoc_1" {
  subnet_id      = aws_subnet.forge_public_subnet.id
  route_table_id = aws_route_table.forge_public_rt.id
}

resource "aws_route_table_association" "forge_public_assoc_2" {
  subnet_id      = aws_subnet.forge_public_subnet_2.id
  route_table_id = aws_route_table.forge_public_rt.id
}

resource "aws_security_group" "forge_ec2_sg" {
  name        = "forge-ec2-sg"
  description = "Allow SSH and HTTP for Forge test"
  vpc_id      = aws_vpc.forge_vpc.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["38.183.10.149/32"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["38.183.10.149/32"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "forge-ec2-sg"
  }
}

resource "aws_instance" "forge_test_ec2" {
  ami           = "ami-07a0715df72e58928"
  instance_type = "t3.micro"

  subnet_id = aws_subnet.forge_public_subnet.id

  vpc_security_group_ids = [
    aws_security_group.forge_ec2_sg.id
  ]

  tags = {
    Name = "forge-test-ec2"
  }
}

resource "aws_ecr_repository" "forge_backend" {
  name = "forge-backend"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Name = "forge-backend"
  }
}