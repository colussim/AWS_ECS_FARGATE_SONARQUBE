![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)![Amazon ECS](https://img.shields.io/static/v1?style=for-the-badge&message=Amazon+ECS&color=222222&logo=Amazon+ECS&logoColor=FF9900&label=)![Static Badge](https://img.shields.io/badge/Go-v1.21-blue:) ![Static Badge](https://img.shields.io/badge/AWS_CDK-v2.115.0-blue:)


# Welcome to your CDK Deployment with Go.

The purpose of this deployment is to created a lambda function to create a sonarqube database.

* The `cdk.json` file tells the CDK toolkit how to execute your app.
* The `Config.json` Contains the parameters to be initialized to deploy the lambda function :
```


config.json :

    SecretName:         secret name primary RDS DB instance        
	Endpoint            database hosts
	Secretmastername    Secret master name
	PortDB              database port
	MasterDB            master db name : postgres
	MasterUser          master db user : postgres
	LambdaFunctionName  Lambda Function Name
	SecretNameSonarqube secret name sonarqube db
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
X : is Index for dbname Index is set in **config_crd.json** file

## The entryparam is the environment variable we use for submitting config options to the lambda:

```
"DATABASE_HOST":     database hosts
"DATABASE_PORT":     database port
"DATABASE_NAME":     master db name : postgres
"DATABASE_USERNAME": master db user
"DATABASE_PASSWORD": master db password
"PASS_SONAR":        password for sonarqube database user
"DATABASE_PARTNER":  index for dbname

```

## ✅ Useful commands

 * `./cdk.sh deploy`      deploy this stack to your default AWS account/region
 * `./cdk.sh destroy`     cleaning up stack

## ✅ Setup Environment

Run the following command to automatically install all the required modules based on the go.mod and go.sum files:

```bash
AWS_ECS_FARGATE_SONARQUBE:/createdb/> go mod download

```
## ✅ Deploying your Lambda function

Let’s deploy a Lambda function! When you’re ready, run **cdk.sh deploy**

```bash
AWS_ECS_FARGATE_SONARQUBE:/createdb/> ./cdk.sh deploy

Lambdatack02: deploying... [1/1]
Lambdatack02: creating CloudFormation changeset...

 ✅  Lambdatack02

✨  Deployment time: 181.89s

Stack ARN:
arn:aws:cloudformation:eu-central-1:xxxxxxxxxxxxxx:stack/Lambdatack02/08b175c0-9da0-11ee-b6d9-0611aa229bc7

✨  Total time: 195.44s
```

-----
<table>
<tr style="border: 0px transparent">
	<td style="border: 0px transparent"> <a href="../database/README.md" title="Creating AWS RDS instance">⬅ Previous</a></td><td style="border: 0px transparent"><a href="../sonarqube/README.md" title="Deploy SonarQube">Next  ➡</a></td><td style="border: 0px transparent"><a href="../README.md" title="home">🏠</a></td>
</tr>
<tr style="border: 0px transparent">
<td style="border: 0px transparent">Creating AWS RDS instance</td><td style="border: 0px transparent">Deploy SonarQube</td><td style="border: 0px transparent"></td>
</tr>

</table>

