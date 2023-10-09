 ![SonarQube](images/sonar.png)![Amazon ECS](https://img.shields.io/static/v1?style=for-the-badge&message=Amazon+ECS&color=222222&logo=Amazon+ECS&logoColor=FF9900&label=)![Static Badge](https://img.shields.io/badge/Go-v1.21-blue:) ![Static Badge](https://img.shields.io/badge/AWS_CDK-v2.96.2-blue:)



SonarQube is a powerful code quality management tool that helps developers identify and correct code quality and security issues. The purpose of this tutorial is to look at how to deploy SonarQube on AWS Elastic Container Service (ECS) Fargate.
This deployment is the extraction of a larger deployment that included several ECS Fargate servers as well as several sonarqube instances and sonarqube databases on an RDS instance.

![Azure AKS, Azure AKS](/images/config.png)

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
The ports to be authorized for input/output are : 9000 (sonarqube default port) , 2049 (EFS Volume) 

We'll use the same VPC and Security Group to deploy the PostgreSQL RDS instance and our SonarQube workload.

## Steps

### Deploy AWS RDS PostgreSQL.

go to directory database (please read the README.md)

## Deploy AWS ECS Fargate cluster.

go to directory ecs (please read the README.md)

## Deploy AWS Lambda function 

go to directory createdb (please read the README.md)

## Deploy SonarQube

go to directory sonarqube (please read the README.md)



# Next steps
- create a DNS sub-domain
- Create a ssl certification 
- integration in cloudfront
