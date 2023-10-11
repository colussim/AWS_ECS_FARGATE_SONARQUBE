![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)![Amazon ECS](https://img.shields.io/static/v1?style=for-the-badge&message=Amazon+ECS&color=222222&logo=Amazon+ECS&logoColor=FF9900&label=)![Static Badge](https://img.shields.io/badge/Go-v1.21-blue:) ![Static Badge](https://img.shields.io/badge/AWS_CDK-v2.96.2-blue:)


# Welcome to your CDK Deployment with Go.

The purpose of this deployment is to created a lambda function to create a sonarqube database.

* The `cdk.json` file tells the CDK toolkit how to execute your app.
* The `Config.json` Contains the parameters to be initialized to deploy the lambda function :
```
Config.json :

    SecretName:         secret name primary RDS DB instance         
	Region:             Deployment region
	Endpoint            database hosts
	Secretmastername    Secret master name
	PortDB              database port
	MasterDB            master db name : postgres
	MasterUser          master db user : postgres
	Index               index for dbname : 1
	VPCid               the VPC ID
	SecurityGroupID     Security Group ID
	LambdaFunctionName  Lambda Function Name
	SecretNameSonarqube secret name sonarqube db
```

Before deploying your task, you need to modify the createdb.go file and setup environment variables: 

* Set your AWS account number and your deployment region :

```
createdb.go

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String("xxxxxx"),
		Region:  jsii.String("eu-central-1"),
	}
}
``` 

* And build your lambda function (for linux OS)

```
:> cd lambda
:> GOOS=linux GOARCH=amd64 go build -o main main.go 
```

## What does the lambda do?
Basically the lambda does four things:

* Create a sonarqube_X role
* Create a sonarqube_0X database
X : is index for dbname

## The entryparam is the environment variable we use for submitting config options to the lambda:

```
"DATABASE_HOST":     database hosts
"DATABASE_PORT":     database port
"DATABASE_NAME":     master db name : postgres
"DATABASE_USERNAME": master db user
"DATABASE_PASSWORD": master db password
"DATABASE_PARTNER":  index for dbname

## Useful commands

 * `cdk deploy`      deploy this stack to your default AWS account/region
 * `cdk diff`        compare deployed stack with current state
 * `cdk synth`       emits the synthesized CloudFormation template
