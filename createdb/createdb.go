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

type DeletedbStackProps struct {
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
	SecretName          string
	Endpoint            string
	Secretmastername    string
	PortDB              string
	MasterDB            string
	MasterUser          string
	LambdaFunctionName  string
	SecretNameSonarqube string
}

type SecretData struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
		fmt.Println("❌ Error Get Secret RDS DB", err)
		os.Exit(1)

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

	key2 := "dbPasswordpart"
	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(Secretname),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		fmt.Println("❌ Error Get Secret Sonar", err)
		os.Exit(1)

	}

	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString

	var secretData map[string]interface{}
	json.Unmarshal([]byte(secretString), &secretData)

	// Get Master User/Password for RDS Instance

	Pass := fmt.Sprint(secretData[key2])

	return Pass
}

func CreateLambdaFn(stack awscdk.Stack, PartVpc awsec2.IVpc, sonarSG awsec2.ISecurityGroup, AppConfig Configuration, User string, Pass string, PassSonar string, LambdaFunctionName string, Index string) awslambda.IFunction {

	lambdaFn := awslambda.NewFunction(stack, &AppConfig.LambdaFunctionName, &awslambda.FunctionProps{
		Runtime:           awslambda.Runtime_PROVIDED_AL2(), // old version awslambda.Runtime_GO_1_X(),
		Code:              awslambda.Code_FromAsset(jsii.String("lambda/"), &awss3assets.AssetOptions{}),
		Handler:           jsii.String("main"),
		Vpc:               PartVpc,
		FunctionName:      &LambdaFunctionName,
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
			"DATABASE_PARTNER":  &Index,
			"PASS_SONAR":        &PassSonar,
		},
	})
	return lambdaFn
}

// Check if a Lambda function exists
func checkLambdaFunctionExists(ctx context.Context, client *lambda.Client, functionName string) (bool, error) {
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	}
	_, err := client.GetFunction(ctx, input)

	if err != nil {
		// Lambda function not exists
		fmt.Println("❗️ Lambda function not exists")
		return false, nil
	}
	// Lambda function exists
	fmt.Println("❗️ Lambda function exists")
	return true, nil
}

func NewCreatedbStack1(scope constructs.Construct, id string, props *CreatedbStackProps, AppConfig Configuration, AppConfig1 ConfAuth) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Get Context
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("❌ Error creating AWS session:", err)
		os.Exit(1)
	}

	// Get Security Group
	sonarSG := awsec2.SecurityGroup_FromLookupById(stack, jsii.String("SG"), &AppConfig1.SecurityGroupID)
	//sonarSG := AppConfig.SecurityGroupID
	// Get VPC
	PartVpc := awsec2.Vpc_FromLookup(stack, &AppConfig1.VPCid, &awsec2.VpcLookupOptions{VpcId: &AppConfig1.VPCid})

	lambdaClient := lambda.NewFromConfig(config)

	// Specify the name of the Lambda function you want to check
	functionName := AppConfig.LambdaFunctionName + AppConfig1.Index

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
			fmt.Println("❌ Error getting Lambda function configuration: %v", err)
			os.Exit(1)

		}

		// Copy existing environment variables
		newEnvVars := make(map[string]string)
		for k, v := range configOutput.Environment.Variables {
			newEnvVars[k] = v
		}
		newEnvVars[key] = AppConfig1.Index

		updateInput := &lambda.UpdateFunctionConfigurationInput{
			FunctionName: aws.String(functionName),
			//Environment:  &types.Environment{&key: AppConfig.Index},
			Environment: &types.Environment{
				Variables: newEnvVars,
			},
		}

		_, updateErr := lambdaClient.UpdateFunctionConfiguration(context.TODO(), updateInput)
		if updateErr != nil {
			fmt.Println("❌ Error updating Lambda function configuration: %v", err)
			os.Exit(1)

		}

		fmt.Printf("✅ Environment variable %s updated for Lambda function %s\n", key, functionName)

		input3 := &lambda.InvokeInput{
			FunctionName: aws.String(functionName),
		}

		resultex, err := lambdaClient.Invoke(context.TODO(), input3)
		if err != nil {
			panic(err)
		} else {
			fmt.Printf("✅ Lambda function executed%s :\n", &resultex)
		}

	} else {
		RDSsecret := AppConfig.SecretName + AppConfig1.Index
		SonarSecret := AppConfig.SecretNameSonarqube + AppConfig1.Index
		fmt.Printf("✅ Lambda function %s does not exist\n", functionName)
		User, Pass := GetSecret(RDSsecret, config)
		PassSonar := GetSecret1(SonarSecret, config)
		CreateLambdaFn(stack, PartVpc, sonarSG, AppConfig, User, Pass, PassSonar, functionName, AppConfig1.Index)
	}

	return stack
}

func NewCreatedbStack(scope constructs.Construct, id string, props *CreatedbStackProps, AppConfig Configuration, AppConfig1 ConfAuth) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Get Context
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("❌ Error creating AWS session:", err)
		os.Exit(1)
	}

	// Get Security Group
	sonarSG := awsec2.SecurityGroup_FromLookupById(stack, jsii.String("SG"), &AppConfig1.SecurityGroupID)

	// Get VPC
	PartVpc := awsec2.Vpc_FromLookup(stack, &AppConfig1.VPCid, &awsec2.VpcLookupOptions{VpcId: &AppConfig1.VPCid})

	lambdaClient := lambda.NewFromConfig(config)

	// Specify the name of the Lambda function you want to check
	functionName := AppConfig.LambdaFunctionName + AppConfig1.Index

	functionExists, err := checkLambdaFunctionExists(context.TODO(), lambdaClient, functionName)
	if err != nil {
		log.Fatalf("Error checking Lambda function existence: %v", err)
	}

	if !functionExists {
		// If the function does not exist, get the secrets and create the Lambda function
		RDSsecret := AppConfig.SecretName + AppConfig1.Index
		SonarSecret := AppConfig.SecretNameSonarqube + AppConfig1.Index

		User, Pass := GetSecret(RDSsecret, config)
		PassSonar := GetSecret1(SonarSecret, config)

		lambdaFunction := CreateLambdaFn(stack, PartVpc, sonarSG, AppConfig, User, Pass, PassSonar, functionName, AppConfig1.Index)
		if lambdaFunction != nil {
			fmt.Printf("✅ Lambda function %s created successfully\n", functionName)
		} else {
			fmt.Printf("❌ Failed to create Lambda function %s\n", functionName)
		}

	} else {
		// If the function exists, log information and update
		fmt.Printf("Lambda function %s exists\n", functionName)

		functionName := AppConfig.LambdaFunctionName
		key := "DATABASE_PARTNER"

		// Get the current Lambda function configuration
		ConfigFn := &lambda.GetFunctionConfigurationInput{
			FunctionName: aws.String(functionName),
		}

		configOutput, err := lambdaClient.GetFunctionConfiguration(context.TODO(), ConfigFn)
		if err != nil {
			fmt.Println("❌ Error getting Lambda function configuration: %v", err)
			os.Exit(1)

		}

		// Copy existing environment variables
		newEnvVars := make(map[string]string)
		for k, v := range configOutput.Environment.Variables {
			newEnvVars[k] = v
		}
		newEnvVars[key] = AppConfig1.Index

		updateInput := &lambda.UpdateFunctionConfigurationInput{
			FunctionName: aws.String(functionName),
			//Environment:  &types.Environment{&key: AppConfig.Index},
			Environment: &types.Environment{
				Variables: newEnvVars,
			},
		}

		_, updateErr := lambdaClient.UpdateFunctionConfiguration(context.TODO(), updateInput)
		if updateErr != nil {
			fmt.Println("❌ Error updating Lambda function configuration: %v", err)
			os.Exit(1)

		}

		fmt.Printf("✅ Environment variable %s updated for Lambda function %s\n", key, functionName)

		input3 := &lambda.InvokeInput{
			FunctionName: aws.String(functionName),
		}

		resultex, err := lambdaClient.Invoke(context.TODO(), input3)
		if err != nil {
			panic(err)
		} else {
			fmt.Printf("✅ Lambda function executed%s :\n", &resultex)
		}

	}

	return stack
}

func DeleteCreatedbStack(scope constructs.Construct, id string, props *DeletedbStackProps, AppConfig Configuration, AppConfig1 ConfAuth) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	os.Setenv("AWS_SDK_LOAD_CONFIG", "true")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("❌ Failed to load AWS configuration", err)
		os.Exit(1)
	}

	functionName := AppConfig.LambdaFunctionName + AppConfig1.Index

	client := lambda.NewFromConfig(cfg)

	// Call DeleteFunction API to delete the Lambda function
	_, err = client.DeleteFunction(context.TODO(), &lambda.DeleteFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		fmt.Println("❌ Failed to delete Lambda function", err)
		os.Exit(1)
	}

	awscdk.NewCfnOutput(stack, jsii.String("Lambda function :"), &awscdk.CfnOutputProps{
		Value: aws.String("Deleted"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	// Read configuration from config.json file
	var configcrd ConfAuth
	var config1 Configuration
	var AppConfig1, AppConfig = GetConfig(configcrd, config1)
	Stack1 := "Lambdatack" + AppConfig1.Index
	Stack2 := "DeleteLambdatack" + AppConfig1.Index

	app := awscdk.NewApp(nil)

	destroy := app.Node().TryGetContext(jsii.String("destroy"))
	//destroyStr := destroy.(string)
	if destroy == "true" {

		DeleteCreatedbStack(app, Stack2, &DeletedbStackProps{
			awscdk.StackProps{
				Env: env(AppConfig1.Region, AppConfig1.Account),
			},
		}, AppConfig, AppConfig1)

	} else {

		NewCreatedbStack(app, Stack1, &CreatedbStackProps{
			awscdk.StackProps{
				Env: env(AppConfig1.Region, AppConfig1.Account),
			},
		}, AppConfig, AppConfig1)
	}
	app.Synth(nil)
}

func env(Region1 string, Account1 string) *awscdk.Environment {

	return &awscdk.Environment{
		Account: &Account1,
		Region:  &Region1,
	}
}
