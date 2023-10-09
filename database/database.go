package main

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DatabaseStackProps struct {
	awscdk.StackProps
}

// Declare a struct for Config fields
type Configuration struct {
	SecretName            string
	DescSecret            string
	Region                string
	Instanceclass         string
	Version               string
	DBsize                string
	Engine                string
	BackupRetentionPeriod float64
	DBName                string
	DBid                  string
	MasterUsername        string
	MasterUserPassword    string
	SubnetGroup           string
	VPCid                 string
	VPCname               string
	SecurityGroupID       string
	SecretNameSonarqube   string
	SonarqubeUserPassword string
	SonarqubeDescSecret   string
}

type DatabaseConfig struct {
	Username string
	Password string
}

func GetConfig(configjs Configuration) Configuration {

	fconfig, err := os.ReadFile("config.json")
	if err != nil {
		panic("Problem with the configuration file : Config.json")
		os.Exit(1)
	}
	json.Unmarshal(fconfig, &configjs)
	return configjs
}

func NewDatabaseStack(scope constructs.Construct, id string, props *DatabaseStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	var config1 Configuration
	var AppConfig = GetConfig(config1)

	// Create a Secret associated with each database role
	awssecretsmanager.NewSecret(stack, &AppConfig.SecretNameSonarqube, &awssecretsmanager.SecretProps{
		Description: &AppConfig.SonarqubeDescSecret,
		SecretName:  &AppConfig.SecretNameSonarqube,
		SecretObjectValue: &map[string]awscdk.SecretValue{
			"dbPasswordpart": awscdk.SecretValue_UnsafePlainText(jsii.String(AppConfig.SonarqubeUserPassword)),
		},
	})

	// Create a Secret associated with primary RDS DB instance: BBpartner-X
	secret := awssecretsmanager.NewSecret(stack, &AppConfig.SecretName, &awssecretsmanager.SecretProps{
		Description: &AppConfig.DescSecret,
		SecretName:  &AppConfig.SecretName,
		SecretObjectValue: &map[string]awscdk.SecretValue{
			"username": awscdk.SecretValue_UnsafePlainText(jsii.String(AppConfig.MasterUsername)),
			"password": awscdk.SecretValue_UnsafePlainText(jsii.String(AppConfig.MasterUserPassword)),
		},
	})

	// Get SecurityGroup
	securityGroupId := []*string{&AppConfig.SecurityGroupID}
	// Set DbInstanceIdentifier
	DBIDDeploy := AppConfig.DBName + "-" + AppConfig.DBid

	awsrds.NewCfnDBInstance(stack, &DBIDDeploy, &awsrds.CfnDBInstanceProps{
		AllocatedStorage:      &AppConfig.DBsize,
		DbInstanceIdentifier:  &DBIDDeploy,
		DbInstanceClass:       &AppConfig.Instanceclass,
		BackupRetentionPeriod: &AppConfig.BackupRetentionPeriod,
		AvailabilityZone:      &AppConfig.Region,
		Engine:                &AppConfig.Engine,
		EngineVersion:         &AppConfig.Version,
		MasterUsername:        &AppConfig.MasterUsername,
		MasterUserPassword:    &AppConfig.MasterUserPassword,
		MasterUserSecret: &awsrds.CfnDBInstance_MasterUserSecretProperty{
			SecretArn: secret.SecretArn(),
		},
		VpcSecurityGroups: &securityGroupId,
		DbSubnetGroupName: &AppConfig.SubnetGroup,
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewDatabaseStack(app, "DatabaseStack", &DatabaseStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String("xxxxx"),
		Region:  jsii.String("eu-central-1"),
	}
}
