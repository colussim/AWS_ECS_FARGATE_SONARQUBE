![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)![Amazon ECS](https://img.shields.io/static/v1?style=for-the-badge&message=Amazon+ECS&color=222222&logo=Amazon+ECS&logoColor=FF9900&label=)![Static Badge](https://img.shields.io/badge/Go-v1.21-blue:) ![Static Badge](https://img.shields.io/badge/AWS_CDK-v2.96.2-blue:)


# Welcome to your CDK Go project!

The purpose of this deployment is to run an ECS Fargate cluster.

* The `cdk.json` file tells the CDK toolkit how to execute your app.
* The `Config.json` Contains the parameters to be initialized to deploy the task :
```
Config.json :

    Region          Deployment region
	ClusterName     ECS Cluster name
	Index           index for Cluster Name
	VPCid           the VPC ID
	SecurityGroupID Security Group ID
```


Before deploying your task, you need to modify the ecs.go file and setup environment variables: 

Set your AWS account number and your deployment region :

```
ecs.go

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String("xxxxxxx"),
		Region:  jsii.String("eu-central-1"),
	}
}
``` 

## What does this task do?

- Deploy ECS Fargate Cluster

## Useful commands

 * `cdk deploy`      deploy this stack to your default AWS account/region
 * `cdk diff`        compare deployed stack with current state
 * `cdk synth`       emits the synthesized CloudFormation template
