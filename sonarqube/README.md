![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)![Amazon ECS](https://img.shields.io/static/v1?style=for-the-badge&message=Amazon+ECS&color=222222&logo=Amazon+ECS&logoColor=FF9900&label=)![Static Badge](https://img.shields.io/badge/Go-v1.21-blue:) ![Static Badge](https://img.shields.io/badge/AWS_CDK-v2.96.2-blue:)


# Welcome to your CDK Deployment with Go.

* The `cdk.json` file tells the CDK toolkit how to execute your app.
* The `Config.json` Contains the parameters to be initialized to deploy the task :
```
Config.json :

    SecretNameSonarqube secret name to store access sonarqube
	Region              Deployment region
	Endpoint            string
	Secretmastername    Secret master name
	SecretRDS           secret name for RDS database access
	PortDB              database hosts
	MasterDB            Master DB name default postgres
	MasterUser          Master user name default postgres
	Index               index for dbname : 01
	SonarvolumeName1    EFS Volume Data
	SonarvolumeName2    EFS Volume Log
	ClusterName         ECS Cluster name
	ClusterARN          ECS Cluster ARN
	VPCid               the VPC ID
	SecurityGroupID     Security Group ID
	SonarImages         Sonar docker image
	Urlconnect          jdbc url
	UriDatabaseType     postgres
	DesiredCount        count deployment
	Cpu                 CPU number
	MemoryLimitMiB      Memory size
	Taskname            Task name
	EcsRole             ECS Role used for deployment
        RessourcesARNMT     Ressource ARN for EFS : "arn:aws:elasticfilesystem:*:XXXXXXXXX:file-system/*" 

For RessourcesARNMT variable replace XXXXXXXXX with your AWS account.
Before deploying your task, you need to modify the sonarqube.go file and setup environment variables: 

Set your AWS account number and your deployment region :

```
sonarqube.go

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String("xxxxxx"),
		Region:  jsii.String("eu-central-1"),
	}
}
``` 

## What does this task do?

- Provision of 2 EFS volumes
- Create a task in ECS Fargate cluster
- Deploy sonarqube
- Provision an external ip address : to connect to sonarqube


## Useful commands

 * `cdk deploy`      deploy this stack to your default AWS account/region
 * `cdk diff`        compare deployed stack with current state
 * `cdk synth`       emits the synthesized CloudFormation template

## Setup Environment

Run the following command to automatically install all the required modules based on the go.mod and go.sum files:

```bash
:> go mod download
```
