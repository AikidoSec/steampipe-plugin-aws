---
title: "Steampipe Table: aws_api_gateway_vpc_link - Query AWS API Gateway VPC Links using SQL"
description: "Allows users to query AWS API Gateway VPC Links to retrieve the network load balancers targeted by each private integration."
folder: "API Gateway"
---

# Table: aws_api_gateway_vpc_link - Query AWS API Gateway VPC Links using SQL

An AWS API Gateway VPC link connects a REST API to resources running inside an Amazon Virtual Private Cloud without exposing them to the internet. A REST API integration of connection type `VPC_LINK` references a VPC link by its ID, and the VPC link forwards traffic to one or more network load balancers in the VPC.

## Table Usage Guide

The `aws_api_gateway_vpc_link` table in Steampipe provides you with information about each VPC link in a region for API Gateway REST APIs. This table allows you as a DevOps or security engineer to query VPC link details, including the link identifier, status, and the ARNs of the targeted network load balancers. You can utilize this table to resolve which load balancer a private integration actually reaches, since the integration URI host name is not used for routing on a `VPC_LINK` connection.

## Examples

### Basic info
Explore the VPC links in your account to understand which private integrations are available and whether they are ready to serve traffic.

```sql+postgres
select
  id,
  name,
  status,
  target_arns,
  region
from
  aws_api_gateway_vpc_link;
```

```sql+sqlite
select
  id,
  name,
  status,
  target_arns,
  region
from
  aws_api_gateway_vpc_link;
```

### List VPC links that are not available
Identify VPC links that are pending, deleting, or failed, as these prevent API deployments from routing traffic to the VPC.

```sql+postgres
select
  id,
  name,
  status,
  status_message
from
  aws_api_gateway_vpc_link
where
  status <> 'AVAILABLE';
```

```sql+sqlite
select
  id,
  name,
  status,
  status_message
from
  aws_api_gateway_vpc_link
where
  status != 'AVAILABLE';
```

### Get the network load balancer targeted by each VPC link
Determine the load balancer behind each private integration by expanding the target ARNs.

```sql+postgres
select
  l.id,
  l.name,
  t as target_arn
from
  aws_api_gateway_vpc_link as l,
  jsonb_array_elements_text(l.target_arns) as t;
```

```sql+sqlite
select
  l.id,
  l.name,
  t.value as target_arn
from
  aws_api_gateway_vpc_link as l,
  json_each(l.target_arns) as t;
```
