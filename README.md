 ![SonarQube](images/sonar.png)![Amazon ECS](https://img.shields.io/static/v1?style=for-the-badge&message=Amazon+ECS&color=222222&logo=Amazon+ECS&logoColor=FF9900&label=)![Static Badge](https://img.shields.io/badge/Go-v1.21-blue:) ![Static Badge](https://img.shields.io/badge/AWS_CDK-v2.96.2-blue:)



SonarQube is a powerful code quality management tool that helps developers identify and correct code quality and security issues.
This tutorial aims to show you how to set up SonarQube on AWS Elastic Container Service (ECS) Fargate. Throughout this guide, we'll walk you through the steps of deploying SonarQube in an ECS Fargate environment using AWS CDK with Golang.


This deployment is the extraction of a larger deployment that included several ECS Fargate servers as well as several sonarqube instances and sonarqube databases on an RDS instance.

![Azure AKS, Azure AKS](/images/aws-ecs-fargate-sonar.jpg)

## Prerequisites

Before you get started, you’ll need to have these things:

* AWS account
* [AWS Cloud Development Kit (AWS CDK) v2](https://docs.aws.amazon.com/cdk/v2/guide/getting_started.html)
* [Go language installed](https://go.dev/)
* [Node.jjs installed](https://nodejs.org/en)
* A AWS VPC
* A AWS Security Group

When setting up a new AWS environment for our project, one of the first things you'll need to do is create a VPC.
When setting up the VPC, it is essential to configure security groups to control inbound and outbound traffic to and from the VPC. Security groups act as virtual firewalls, allowing only authorized traffic to pass through.
The ports to be authorized (defined in the Security Groups) for input/output are : 9000 (sonarqube default port) , 2049 (EFS Volume) 

We'll use the same VPC and Security Group to deploy the PostgreSQL RDS instance and our SonarQube workload.

## Steps

### Deploy AWS RDS PostgreSQL.

go to directory [database](https://github.com/colussim/AWS_ECS_FARGATE_SONARQUBE/tree/main/database) (please read the README.md)

## Deploy AWS ECS Fargate cluster.

go to directory [ecs](https://github.com/colussim/AWS_ECS_FARGATE_SONARQUBE/tree/main/ecs) (please read the README.md)

## Deploy AWS Lambda function 

go to directory [createdb](https://github.com/colussim/AWS_ECS_FARGATE_SONARQUBE/tree/main/createdb) (please read the README.md)

## Deploy SonarQube

go to directory [sonarqube](https://github.com/colussim/AWS_ECS_FARGATE_SONARQUBE/tree/main/sonarqube) (please read the README.md)



# Next steps

At this stage the sonarqube deployment provides a public address ip, in a next step I'll do the integration in a dns domain and https access with certificate.

- create a DNS sub-domain
- Create a ssl certification 
- integration in cloudfront
