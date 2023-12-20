![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)![Amazon ECS](https://img.shields.io/static/v1?style=for-the-badge&message=Amazon+ECS&color=222222&logo=Amazon+ECS&logoColor=FF9900&label=)![Static Badge](https://img.shields.io/badge/Go-v1.21-blue:) ![Static Badge](https://img.shields.io/badge/AWS_CDK-v2.115.0-blue:)


# Welcome to your CDK Deployment with Go.

The purpose of this deployment is to run an AWS RDS PostgreSQL instance.


* The `cdk.json` file tells the CDK toolkit how to execute your app.

## What does this task do?

- ECS Task Execution Role

## ✅ Useful commands

 * `cdk deploy`      deploy this stack to your default AWS account/region
 * `cdk destroy --force`     cleaning up stack

## Setup Environment

Run the following command to automatically install all the required modules based on the go.mod and go.sum files:

```bash
AWS_ECS_FARGATE_SONARQUBE:/deployrole/> go mod download

```
## ✅ Deploying your RDS instance

Let’s deploy a RDS database! When you’re ready, run **cdk deploy**

```bash
AWS_ECS_FARGATE_SONARQUBE:/deployrole/> ./cdk.sh deploy


DeployRole02: creating CloudFormation changeset...

 ✅  DeployRole02

✨  Deployment time: 32.17s

Stack ARN:
arn:aws:cloudformation:eu-central-1:xxxxxxxxxx:stack/DeployRole02/e99cd380-9f3b-11ee-967c-060ac76d9ab3

✨  Total time: 34.06s

```

<table>
<tr style="border: 0px transparent">
	<td style="border: 0px transparent"><a href="../README.md" title="home">🏠</a></td><td style="border: 0px transparent"><a href="../ecs/README.md" title="Deploy AWS ECS Fargate cluster">Next ➡</a></td>
</tr>
<tr style="border: 0px transparent">
<td style="border: 0px transparent">Introduction</td><td style="border: 0px transparent">AWS ECS Fargate cluster</td><td style="border: 0px transparent"></td>
</tr>

</table>
