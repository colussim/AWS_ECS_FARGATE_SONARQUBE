package main

import (
	"encoding/json"
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

// Declare a struct for Config fields
type Configuration struct {
	Region          string
	ClusterName     string
	Index           string
	VPCid           string
	SecurityGroupID string
}

func GetConfig(configjs Configuration) Configuration {

	fconfig, err := os.ReadFile("config.json")
	if err != nil {
		panic("Problem with the configuration file : config.json")
		os.Exit(1)
	}
	json.Unmarshal(fconfig, &configjs)
	return configjs
}

func NewEcsStack(scope constructs.Construct, id string, props *EcsStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	var config1 Configuration
	var AppConfig = GetConfig(config1)

	// Get VPC
	PartVpc := awsec2.Vpc_FromLookup(stack, &AppConfig.VPCid, &awsec2.VpcLookupOptions{VpcId: &AppConfig.VPCid})

	// Create Cluster
	clustername := AppConfig.ClusterName + AppConfig.Index
	awsecs.NewCluster(stack, jsii.String("FargateCluster"), &awsecs.ClusterProps{
		Vpc:         PartVpc,
		ClusterName: &clustername,
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewEcsStack(app, "EcsStack", &EcsStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {

	return &awscdk.Environment{
		Account: jsii.String("xxxxxxx"),
		Region:  jsii.String("eu-central-1"),
	}

}
