#!/bin/bash
set -e

# Initialize Terraform
terraform init

# Plan the deployment
terraform plan -out=tfplan

# Apply the deployment
terraform apply tfplan

# Display the Redis connection information
echo "Redis deployment complete!"
echo "----------------------------------------"
echo "Redis Connection String: $(terraform output -raw redis_connection_string)"
echo "Redis Endpoint: $(terraform output -raw redis_endpoint)"
echo "Redis Port: $(terraform output -raw redis_port)"
echo "----------------------------------------"
echo "Add the Redis URL to your environment:"
echo "export REDIS_URL=\"$(terraform output -raw redis_connection_string)\""
echo ""
echo "If running locally, add this to your .env file or shell startup script."
echo "If using a deployment platform, add this as an environment variable."