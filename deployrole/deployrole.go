package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DeployroleStackProps struct {
	awscdk.StackProps
}

// Declare a struct for Config fields
type Configuration struct {
	RoleName string
}

type ConfAuth struct {
	Region          string
	Account         string
	SSOProfile      string
	Index           string
	VPCid           string
	SecurityGroupID string
}

func GetConfig(configcrd ConfAuth, configjs Configuration) (ConfAuth, Configuration) {

	fconfig, err := os.ReadFile("config.json")
	if err != nil {
		panic("❌ Problem with the configuration file : config.json")
		os.Exit(1)
	}
	if err := json.Unmarshal(fconfig, &configjs); err != nil {
		fmt.Println("❌ Error unmarshaling JSON:", err)
		os.Exit(1)
	}

	fconfig2, err := os.ReadFile("../config_crd.json")
	if err != nil {
		panic("❌ Problem with the configuration file : config_crd.json")
		os.Exit(1)
	}
	if err := json.Unmarshal(fconfig2, &configcrd); err != nil {
		fmt.Println("❌ Error unmarshaling JSON:", err)
		os.Exit(1)
	}
	return configcrd, configjs
}

func NewDeployroleStack(scope constructs.Construct, id string, props *DeployroleStackProps, AppConfig Configuration, AppConfig1 ConfAuth) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Set Variables
	var AdmRole = AppConfig.RoleName + AppConfig1.Index

	// ARN policies for Role eksadmin
	var policyArn2 = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
	var policyArn3 = "arn:aws:iam::aws:policy/AmazonECS_FullAccess"
	var policyArn4 = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"

	// Define the trusted service principals
	trustedPrincipals := awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil)

	// Define IAM role for the EKS cluster.
	CDKAdminRole := awsiam.NewRole(stack, &AdmRole, &awsiam.RoleProps{
		AssumedBy: trustedPrincipals,
		RoleName:  &AdmRole,
	})

	CDKAdminRole.AddManagedPolicy(awsiam.ManagedPolicy_FromManagedPolicyArn(stack, jsii.String("AmazonEC2ContainerServiceforEC2Role"), &policyArn2))
	CDKAdminRole.AddManagedPolicy(awsiam.ManagedPolicy_FromManagedPolicyArn(stack, jsii.String("AmazonECS_FullAccess"), &policyArn3))
	CDKAdminRole.AddManagedPolicy(awsiam.ManagedPolicy_FromManagedPolicyArn(stack, jsii.String("AmazonECSTaskExecutionRolePolicy"), &policyArn4))

	return stack
}

func main() {
	defer jsii.Close()
	var configcrd ConfAuth
	var config1 Configuration
	var AppConfig1, AppConfig = GetConfig(configcrd, config1)
	Stack := "DeployRole" + AppConfig1.Index

	app := awscdk.NewApp(nil)

	NewDeployroleStack(app, Stack, &DeployroleStackProps{
		awscdk.StackProps{
			Env: env(AppConfig1.Region, AppConfig1.Account),
		},
	}, AppConfig, AppConfig1)

	app.Synth(nil)
}

func env(Region1 string, Account1 string) *awscdk.Environment {

	return &awscdk.Environment{
		Account: &Account1,
		Region:  &Region1,
	}
}
