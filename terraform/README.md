# Snagbot Redis Deployment

This directory contains Terraform configurations to deploy a Redis instance for SnagBot on AWS ElastiCache.

## Prerequisites

1. [Terraform](https://www.terraform.io/) installed
2. AWS CLI configured with appropriate credentials
3. Basic understanding of AWS services

## Deployment Steps

1. Configure AWS CLI with appropriate credentials:
   ```
   aws configure
   ```

2. Review and modify `variables.tf` if needed:
   - Change the default AWS region
   - Adjust the environment name
   - Change Redis instance type if needed

3. Run the deployment script:
   ```
   ./deploy.sh
   ```

4. The script will:
   - Initialize Terraform
   - Plan the deployment
   - Apply the changes after confirmation
   - Display the Redis connection information

5. Add the Redis URL to your application environment:
   ```
   export REDIS_URL="redis://your-redis-endpoint:6379"
   ```

## Configuration Variables

- `aws_region`: AWS region to deploy resources (default: us-west-2)
- `environment`: Environment name (default: dev)
- `redis_node_type`: ElastiCache Redis node type (default: cache.t3.micro)
- `redis_port`: Redis port (default: 6379)

## Cleanup

To remove all resources created by this Terraform config:

```
terraform destroy
```

## Security Notes

- The default security group allows Redis access from anywhere (0.0.0.0/0)
- For production deployments, restrict this to your application's IP or VPC CIDR