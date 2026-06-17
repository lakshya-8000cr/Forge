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