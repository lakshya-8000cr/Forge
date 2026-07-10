resource "aws_eks_cluster" "forge_eks" {
  name     = "forge-eks"
  role_arn = aws_iam_role.eks_cluster_role.arn

  vpc_config {
    subnet_ids = [
      aws_subnet.forge_public_subnet.id,
      aws_subnet.forge_public_subnet_2.id
    ]
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_policy
  ]

  tags = {
    Name = "forge-eks"
  }
}


resource "aws_eks_node_group" "forge_nodes" {

  cluster_name    = aws_eks_cluster.forge_eks.name

  node_group_name = "forge-workers"

  node_role_arn   = aws_iam_role.eks_node_role.arn

  subnet_ids = [

    aws_subnet.forge_public_subnet.id,

    aws_subnet.forge_public_subnet_2.id
  ]

  scaling_config {

    desired_size = 2

    min_size = 1

    max_size = 3
  }

  instance_types = [

    "t3.small"
  ]

  depends_on = [

    aws_iam_role_policy_attachment.worker_node_policy,

    aws_iam_role_policy_attachment.cni_policy,

    aws_iam_role_policy_attachment.ecr_read_policy
  ]

  tags = {

    Name = "forge-workers"
  }
}
resource "aws_instance" "forge_k6_runner" {
  ami           = "ami-07a0715df72e58928"
  instance_type = "t3.small"
  subnet_id     = aws_subnet.forge_public_subnet.id

  vpc_security_group_ids = [
    aws_security_group.forge_ec2_sg.id
  ]

  # 🔑 NAYI KEY YAHAN ADD KARO (Ensure properties reflect this exact name)
  key_name = "forge-key-v2" 

  user_data = <<-EOF
              #!/bin/bash
              sudo apt update -y
              sudo apt install gnupg software-properties-common curl -y
              curl -s https://dl.k6.io/key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/k6-archive-keyring.gpg
              echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
              sudo apt update -y
              sudo apt install k6 -y
              EOF

  tags = {
    Name = "forge-test-ec2"
  }
}