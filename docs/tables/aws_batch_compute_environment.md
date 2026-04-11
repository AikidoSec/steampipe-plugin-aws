---
title: "Steampipe Table: aws_batch_compute_environment - Query AWS Batch Compute Environments using SQL"
description: "Allows users to query AWS Batch Compute Environments for detailed configuration, status, and capacity information."
folder: "Batch"
---

# Table: aws_batch_compute_environment

AWS Batch compute environments provide the compute capacity where AWS Batch jobs run. They define the backing infrastructure, such as EC2, Spot, Fargate, or EKS-based compute resources, along with scaling and networking settings.

## Table Usage Guide

The `aws_batch_compute_environment` table provides insights into AWS Batch compute environments. Use it to inventory compute backends, inspect networking and capacity settings, and review state and status across regions.

## Examples

### Basic info

```sql+postgres
select
  compute_environment_name,
  type,
  state,
  status
from
  aws_batch_compute_environment;
```

```sql+sqlite
select
  compute_environment_name,
  type,
  state,
  status
from
  aws_batch_compute_environment;
```

### Show compute environments and service roles

```sql+postgres
select
  compute_environment_name,
  service_role,
  arn
from
  aws_batch_compute_environment;
```

```sql+sqlite
select
  compute_environment_name,
  service_role,
  arn
from
  aws_batch_compute_environment;
```
