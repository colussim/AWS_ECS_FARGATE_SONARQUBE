package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type EcsStackProps struct {
	awscdk.StackProps
}

type ConfAuth struct {
	Region          string
	Account         string
	SSOProfile      string
	Index           string
	VPCid           string
	SecurityGroupID string
}

// Declare a struct for Config fields
type Configuration struct {
	ClusterName string
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

func NewEcsStack(scope constructs.Construct, id string, props *EcsStackProps, AppConfig Configuration, AppConfig1 ConfAuth) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Get VPC
	PartVpc := awsec2.Vpc_FromLookup(stack, &AppConfig1.VPCid, &awsec2.VpcLookupOptions{VpcId: &AppConfig1.VPCid})

	// Create Cluster
	clustername := AppConfig.ClusterName + AppConfig1.Index
	awsecs.NewCluster(stack, jsii.String("FargateCluster"), &awsecs.ClusterProps{
		Vpc:         PartVpc,
		ClusterName: &clustername,
	})

	return stack
}

func main() {
	defer jsii.Close()

	// Read configuration from config.json file
	var configcrd ConfAuth
	var config1 Configuration
	var AppConfig1, AppConfig = GetConfig(configcrd, config1)
	Stack1 := "EcsStack" + AppConfig1.Index

	app := awscdk.NewApp(nil)

	NewEcsStack(app, Stack1, &EcsStackProps{
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
