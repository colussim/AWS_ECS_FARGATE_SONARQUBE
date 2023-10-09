package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CreatedbStackProps struct {
	awscdk.StackProps
}

// Declare a struct for Config fields
type Configuration struct {
	SecretName          string
	Region              string
	Endpoint            string
	Secretmastername    string
	PortDB              string
	MasterDB            string
	MasterUser          string
	Index               string
	VPCid               string
	SecurityGroupID     string
	LambdaFunctionName  string
	SecretNameSonarqube string
}

type SecretData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func GetConfig(configjs Configuration) Configuration {

	fconfig, err := os.ReadFile("Config.json")
	if err != nil {
		panic("Problem with the configuration file : config.json")
		os.Exit(1)
	}
	json.Unmarshal(fconfig, &configjs)
	return configjs
}

func GetSecret(Secretname string, config aws.Config) (string, string) {
	key1 := "username"
	key2 := "password"

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(Secretname),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Fatal(err.Error())

	}

	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString

	var secretData map[string]interface{}
	json.Unmarshal([]byte(secretString), &secretData)

	// Get Master User/Password for RDS Instance
	User := fmt.Sprint(secretData[key1])
	Pass := fmt.Sprint(secretData[key2])

	return User, Pass
}

func GetSecret1(Secretname string, config aws.Config) string {

	key2 := "password"
	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(Secretname),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Fatal(err.Error())

	}

	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString

	var secretData map[string]interface{}
	json.Unmarshal([]byte(secretString), &secretData)

	// Get Master User/Password for RDS Instance

	Pass := fmt.Sprint(secretData[key2])

	return Pass
}

func CreateLambdaFn(stack awscdk.Stack, PartVpc awsec2.IVpc, sonarSG awsec2.ISecurityGroup, AppConfig Configuration, User string, Pass string, PassSonar string) {

	awslambda.NewFunction(stack, &AppConfig.LambdaFunctionName, &awslambda.FunctionProps{
		Runtime:           awslambda.Runtime_GO_1_X(),
		Code:              awslambda.Code_FromAsset(jsii.String("lambda/"), &awss3assets.AssetOptions{}),
		Handler:           jsii.String("main"),
		Vpc:               PartVpc,
		FunctionName:      &AppConfig.LambdaFunctionName,
		AllowPublicSubnet: aws.Bool(true),
		Description:       jsii.String("Init SonarQube Database"),
		SecurityGroups:    &[]awsec2.ISecurityGroup{sonarSG},
		Tracing:           awslambda.Tracing_ACTIVE,
		Timeout:           awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize:        jsii.Number(512),
		Environment: &map[string]*string{
			"DATABASE_HOST":     &AppConfig.Endpoint,
			"DATABASE_PORT":     &AppConfig.PortDB,
			"DATABASE_NAME":     &AppConfig.MasterDB,
			"DATABASE_USERNAME": &User,
			"DATABASE_PASSWORD": &Pass,
			"DATABASE_PARTNER":  &AppConfig.Index,
		},
	})

}

// Check if a Lambda function exists
func checkLambdaFunctionExists(ctx context.Context, client *lambda.Client, functionName string) (bool, error) {
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	}
	_, err := client.GetFunction(ctx, input)

	if err != nil {
		// Lambda function not exists
		return false, nil
	}
	// Lambda function exists
	return true, nil
}

func NewCreatedbStack(scope constructs.Construct, id string, props *CreatedbStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Load configuration
	var config1 Configuration
	var AppConfig = GetConfig(config1)

	// Get Context
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Get Security Group
	sonarSG := awsec2.SecurityGroup_FromLookupById(stack, jsii.String("SG"), &AppConfig.SecurityGroupID)
	//sonarSG := AppConfig.SecurityGroupID
	// Get VPC
	PartVpc := awsec2.Vpc_FromLookup(stack, &AppConfig.VPCid, &awsec2.VpcLookupOptions{VpcId: &AppConfig.VPCid})

	lambdaClient := lambda.NewFromConfig(config)

	// Specify the name of the Lambda function you want to check
	functionName := AppConfig.LambdaFunctionName

	functionExists, err := checkLambdaFunctionExists(context.TODO(), lambdaClient, functionName)

	if err != nil {
		log.Fatalf("Error checking Lambda function existence: %v", err)
	}

	if functionExists {
		fmt.Printf("Lambda function %s exists\n", functionName)

		functionName := AppConfig.LambdaFunctionName
		key := "DATABASE_PARTNER"

		// Get the current Lambda function configuration
		ConfigFn := &lambda.GetFunctionConfigurationInput{
			FunctionName: aws.String(functionName),
		}

		configOutput, err := lambdaClient.GetFunctionConfiguration(context.TODO(), ConfigFn)
		if err != nil {
			log.Fatalf("Error getting Lambda function configuration: %v", err)
		}

		// Copy existing environment variables
		newEnvVars := make(map[string]string)
		for k, v := range configOutput.Environment.Variables {
			newEnvVars[k] = v
		}
		newEnvVars[key] = AppConfig.Index

		updateInput := &lambda.UpdateFunctionConfigurationInput{
			FunctionName: aws.String(functionName),
			//Environment:  &types.Environment{&key: AppConfig.Index},
			Environment: &types.Environment{
				Variables: newEnvVars,
			},
		}

		_, updateErr := lambdaClient.UpdateFunctionConfiguration(context.TODO(), updateInput)
		if updateErr != nil {
			log.Fatalf("Error updating Lambda function configuration: %v", updateErr)
		}

		fmt.Printf("Environment variable %s updated for Lambda function %s\n", key, functionName)

		input3 := &lambda.InvokeInput{
			FunctionName: aws.String(functionName),
		}

		resultex, err := lambdaClient.Invoke(context.TODO(), input3)
		if err != nil {
			panic(err)
		} else {
			fmt.Printf("Lambda function executed%s :\n", &resultex)
		}

	} else {
		fmt.Printf("Lambda function %s does not exist\n", functionName)
		User, Pass := GetSecret(AppConfig.SecretName, config)
		PassSonar := GetSecret1(AppConfig.SecretNameSonarqube, config)
		CreateLambdaFn(stack, PartVpc, sonarSG, AppConfig, User, Pass, PassSonar)
	}

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewCreatedbStack(app, "CreatedbStack", &CreatedbStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {

	return &awscdk.Environment{
		Account: jsii.String("xxxxxx"),
		Region:  jsii.String("eu-central-1"),
	}

}
