package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

//EndPointRDS string

type DatabaseStackProps struct {
	awscdk.StackProps
	RDSEndpoint *string
}

type DatabaseStack struct {
	stack       awscdk.Stack
	RDSEndpoint *string
}

type SecretStackProps struct {
	awscdk.StackProps
	//Endpoint string
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
	SecretName            string
	DescSecret            string
	Instanceclass         string
	Version               string
	DBsize                string
	Engine                string
	BackupRetentionPeriod float64
	DBName                string
	DBport                string
	MasterUsername        string
	MasterUserPassword    string
	SubnetGroup           string
	SecretNameSonarqube   string
	SonarqubeUserPassword string
	SonarqubeDescSecret   string
}

type DatabaseConfig struct {
	Username string
	Password string
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

func updateSecretValue(secretName, key, value string) {
	sess := session.Must(session.NewSession())
	secretsManager := secretsmanager.New(sess)

	// Retrieve the existing secret value
	getSecretValueOutput, err := secretsManager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})

	if err != nil {
		fmt.Println("Error getting secret value:", err)
		return
	}

	// Parse the existing secret JSON
	existingSecretJSON := *getSecretValueOutput.SecretString
	existingSecretMap := make(map[string]string)
	err = json.Unmarshal([]byte(existingSecretJSON), &existingSecretMap)
	if err != nil {
		fmt.Println("Error parsing existing secret JSON:", err)
		return
	}

	// Update the specific key in the existing secret map
	existingSecretMap[key] = value

	// Convert the updated map back to JSON
	updatedSecretJSON, err := json.Marshal(existingSecretMap)
	if err != nil {
		fmt.Println("Error marshaling updated secret map to JSON:", err)
		return
	}

	// Update the secret with the new JSON
	_, err = secretsManager.UpdateSecret(&secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(secretName),
		SecretString: aws.String(string(updatedSecretJSON)),
	})

	if err != nil {
		fmt.Println("Error updating secret value:", err)
		return
	}
}

func SecretStack(scope constructs.Construct, id string, props *SecretStackProps, AppConfig Configuration, AppConfig1 ConfAuth, secretName string, RDSEndpoint string) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	rdsEndpoint := RDSEndpoint

	updateSecretValue(secretName, "host", rdsEndpoint)

	awscdk.NewCfnOutput(stack, jsii.String("RdsEndpoint2"), &awscdk.CfnOutputProps{
		Value: &rdsEndpoint,
	})

	return stack

}

func NewDatabaseStack(scope constructs.Construct, id string, props *DatabaseStackProps, AppConfig Configuration, AppConfig1 ConfAuth) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, &id, &sprops)

	SBGroupRDS := AppConfig.SubnetGroup + AppConfig1.Index

	// Get VPC
	PartVpc := awsec2.Vpc_FromLookup(stack, &AppConfig1.VPCid, &awsec2.VpcLookupOptions{VpcId: &AppConfig1.VPCid})
	// Find public subnets in the VPC
	allSubnets := []awsec2.ISubnet{}

	// Iterate through all subnets in the VPC
	for _, subnet := range *PartVpc.PublicSubnets() {

		allSubnets = append(allSubnets, subnet)

	}
	// Create A RDS SubnetGroup
	subnetGroup := awsrds.NewSubnetGroup(stack, &SBGroupRDS, &awsrds.SubnetGroupProps{
		SubnetGroupName: &SBGroupRDS,
		Vpc:             PartVpc,
		VpcSubnets:      &awsec2.SubnetSelection{Subnets: &allSubnets},
		Description:     jsii.String("RDS DB Subnet Group"),
	})

	SecretNameSonar := AppConfig.SecretNameSonarqube + AppConfig1.Index
	EndPoint := ""

	// Create a Secret associated with database role
	awssecretsmanager.NewSecret(stack, &AppConfig.SecretNameSonarqube, &awssecretsmanager.SecretProps{
		Description: &AppConfig.SonarqubeDescSecret,
		SecretName:  &SecretNameSonar,
		SecretObjectValue: &map[string]awscdk.SecretValue{
			"dbPasswordpart": awscdk.SecretValue_UnsafePlainText(jsii.String(AppConfig.SonarqubeUserPassword)),
		},
	})

	SecretNameRDS := AppConfig.SecretName + AppConfig1.Index
	// Create a Secret associated with primary RDS DB instance: BBpartner-X
	secret := awssecretsmanager.NewSecret(stack, &AppConfig.SecretName, &awssecretsmanager.SecretProps{
		Description: &AppConfig.DescSecret,
		SecretName:  &SecretNameRDS,
		SecretObjectValue: &map[string]awscdk.SecretValue{
			"username": awscdk.SecretValue_UnsafePlainText(jsii.String(AppConfig.MasterUsername)),
			"password": awscdk.SecretValue_UnsafePlainText(jsii.String(AppConfig.MasterUserPassword)),
			"host":     awscdk.SecretValue_UnsafePlainText(jsii.String(EndPoint)),
			"port":     awscdk.SecretValue_UnsafePlainText(jsii.String(AppConfig.DBport)),
		},
	})

	// Get SecurityGroup
	securityGroupId := []*string{&AppConfig1.SecurityGroupID}

	// Set DbInstanceIdentifier
	DBIDDeploy := AppConfig.DBName + "-" + AppConfig1.Index

	RegionA := AppConfig1.Region + "a"

	rdsInstance := awsrds.NewCfnDBInstance(stack, &DBIDDeploy, &awsrds.CfnDBInstanceProps{
		AllocatedStorage:      &AppConfig.DBsize,
		DbInstanceIdentifier:  &DBIDDeploy,
		DbInstanceClass:       &AppConfig.Instanceclass,
		BackupRetentionPeriod: &AppConfig.BackupRetentionPeriod,
		AvailabilityZone:      &RegionA,
		Engine:                &AppConfig.Engine,
		EngineVersion:         &AppConfig.Version,
		MasterUsername:        &AppConfig.MasterUsername,
		MasterUserPassword:    &AppConfig.MasterUserPassword,
		MasterUserSecret: &awsrds.CfnDBInstance_MasterUserSecretProperty{
			SecretArn: secret.SecretArn(),
		},
		VpcSecurityGroups: &securityGroupId,
		DbSubnetGroupName: &SBGroupRDS,
		StorageEncrypted:  aws.Bool(true),
		Port:              &AppConfig.DBport,
	})

	subnetGroup.Node().AddDependency(rdsInstance)

	rdsEndpoint := rdsInstance.AttrEndpointAddress()

	awscdk.NewCfnOutput(stack, jsii.String("RdsEndpoint"), &awscdk.CfnOutputProps{
		Value: rdsEndpoint,
	})
	return stack

}

func main() {
	defer jsii.Close()

	// Read configuration from config.json file
	var configcrd ConfAuth
	var config1 Configuration
	var AppConfig1, AppConfig = GetConfig(configcrd, config1)
	Stack1 := "DatabaseStack" + AppConfig1.Index
	Stack2 := "Secretup" + AppConfig1.Index

	app := awscdk.NewApp(nil)

	Endpoint01 := app.Node().TryGetContext(jsii.String("Endpoint"))

	if Endpoint01 != nil {

		Endpoint02 := Endpoint01.(string)
		secretname := AppConfig.SecretName + AppConfig1.Index

		SecretStack(app, Stack2, &SecretStackProps{
			awscdk.StackProps{
				Env: env(AppConfig1.Region, AppConfig1.Account),
			},
			//Endpoint: Endpoint02
		}, AppConfig, AppConfig1, secretname, Endpoint02)

	} else {

		stack1Props := &DatabaseStackProps{
			StackProps: awscdk.StackProps{
				Env: env(AppConfig1.Region, AppConfig1.Account),
			},
			RDSEndpoint: nil, // set the value accordingly
		}

		NewDatabaseStack(app, Stack1, stack1Props, AppConfig, AppConfig1)

	}
	app.Synth(nil)

}

func env(Region1 string, Account1 string) *awscdk.Environment {

	return &awscdk.Environment{
		Account: &Account1,
		Region:  &Region1,
	}
}
