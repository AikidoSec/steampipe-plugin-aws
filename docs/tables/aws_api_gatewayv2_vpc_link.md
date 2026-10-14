---
title: "Steampipe Table: aws_api_gatewayv2_vpc_link - Query AWS API Gateway Version 2 VPC Links using SQL"
description: "Allows users to query AWS API Gateway Version 2 VPC Links to retrieve the subnets and security groups of each private integration."
folder: "API Gateway"
---

# Table: aws_api_gatewayv2_vpc_link - Query AWS API Gateway Version 2 VPC Links using SQL

An AWS API Gateway Version 2 VPC link connects an HTTP API to resources running inside an Amazon Virtual Private Cloud without exposing them to the internet. Unlike a REST API VPC link, a version 2 VPC link is defined by the subnets and security groups it is placed in, and the HTTP API integration points at the private resource directly.

## Table Usage Guide

The `aws_api_gatewayv2_vpc_link` table in Steampipe provides you with information about each VPC link in a region for API Gateway HTTP APIs. This table allows you as a DevOps or security engineer to query VPC link details, including the link identifier, status, version, and the subnets and security groups attached to it. You can utilize this table to review the network placement of private integrations and to find links that are not ready to serve traffic.

## Examples

### Basic info
Explore the version 2 VPC links in your account to understand which private integrations are available and where they are attached.

```sql+postgres
select
  vpc_link_id,
  name,
  vpc_link_status,
  subnet_ids,
  security_group_ids,
  region
from
  aws_api_gatewayv2_vpc_link;
```

```sql+sqlite
select
  vpc_link_id,
  name,
  vpc_link_status,
  subnet_ids,
  security_group_ids,
  region
from
  aws_api_gatewayv2_vpc_link;
```

### List VPC links that are not available
Identify VPC links that are pending, deleting, or failed, as these prevent HTTP APIs from reaching the VPC.

```sql+postgres
select
  vpc_link_id,
  name,
  vpc_link_status,
  vpc_link_status_message
from
  aws_api_gatewayv2_vpc_link
where
  vpc_link_status <> 'AVAILABLE';
```

```sql+sqlite
select
  vpc_link_id,
  name,
  vpc_link_status,
  vpc_link_status_message
from
  aws_api_gatewayv2_vpc_link
where
  vpc_link_status != 'AVAILABLE';
```

### Get the security group details of each VPC link
Review the inbound reach of each private integration by joining the attached security groups.

```sql+postgres
select
  l.vpc_link_id,
  l.name,
  s.group_id,
  s.group_name,
  s.vpc_id
from
  aws_api_gatewayv2_vpc_link as l,
  jsonb_array_elements_text(l.security_group_ids) as sg
  join aws_vpc_security_group as s on s.group_id = sg;
```

```sql+sqlite
select
  l.vpc_link_id,
  l.name,
  s.group_id,
  s.group_name,
  s.vpc_id
from
  aws_api_gatewayv2_vpc_link as l,
  json_each(l.security_group_ids) as sg
  join aws_vpc_security_group as s on s.group_id = sg.value;
```
