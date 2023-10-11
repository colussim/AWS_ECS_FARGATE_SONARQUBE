![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)![Amazon ECS](https://img.shields.io/static/v1?style=for-the-badge&message=Amazon+ECS&color=222222&logo=Amazon+ECS&logoColor=FF9900&label=)![Static Badge](https://img.shields.io/badge/Go-v1.21-blue:) ![Static Badge](https://img.shields.io/badge/AWS_CDK-v2.96.2-blue:)


# Welcome to your CDK Deployment with Go.

The purpose of this deployment is to run an AWS RDS PostgreSQL instance.


* The `cdk.json` file tells the CDK toolkit how to execute your app.
* The `Config.json` Contains the parameters to be initialized to deploy the task :
```
Config.json :

    SecretName:             secret name for RDS database access
	DescSecret:		        Description : Secret associated with primary RDS DB instance: DBSonar01
	Region:                 Deployment region
	Instanceclass:          Instance class
	Version:		        Version of PostgreSQL
	DBsize:			        DB size
	Engine:                 postgres
	BackupRetentionPeriod:  1
	DBName:                 dbname instance
	DBid:                   index for dbname : 01
	MasterUsername :        default postgres
	MasterUserPassword:     master password,
	SubnetGroup:            the subnet group to use
	VPCid :                 the VPC ID
	VPCname:                the VPC Name
	SecurityGroupID :       Security Group ID
	SecretNameSonarqube :   secret name to store access sonarqube
	SonarqubeUserPassword: sonarqube db password
	SonarqubeDescSecret:   Description Secret associated with Access to PostgreSQL databases partner
```    

Before deploying your task, you need to modify the database.go file and setup environment variables: 

Set your AWS account number and your deployment region :

```
database.go

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String("xxxxxx"),
		Region:  jsii.String("eu-central-1"),
	}
}
``` 

## What does this task do?

- Create a secret RDS DB instance
- create a secret for sonarqube database
- Deploy AWS RDS PostgreSQL instance

## Useful commands

 * `cdk deploy`      deploy this stack to your default AWS account/region
 * `cdk diff`        compare deployed stack with current state
 * `cdk synth`       emits the synthesized CloudFormation template

