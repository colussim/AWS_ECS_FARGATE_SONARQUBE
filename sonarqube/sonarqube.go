package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsefs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	awsv1 "github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	secretsmanagerv1 "github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/ecs"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type SonarqubeStackProps struct {
	awscdk.StackProps
}

type GetipStackProps struct {
	awscdk.StackProps
}

// Declare a struct for Config fields
type Configuration struct {
	SecretNameSonarqube string
	Endpoint            string
	SonarvolumeName1    string
	SonarvolumeName2    string
	ClusterName         string
	SonarImages         string
	Urlconnect          string
	UriDatabaseType     string
	DesiredCount        float64
	Cpu                 float64
	MemoryLimitMiB      float64
	Taskname            string
}

type ConfAuth struct {
	Region          string
	Account         string
	SSOProfile      string
	Index           string
	VPCid           string
	SecurityGroupID string
}

type RegionA struct {
	StringRegion []*string
}

type SecretStruct struct {
	DbPasswordPart string `json:"dbPasswordpart"`
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

func NewSonarqubeStack(scope constructs.Construct, id string, props *SonarqubeStackProps, AppConfig Configuration, AppConfig1 ConfAuth) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Get Security Group
	sonarSG := awsec2.SecurityGroup_FromLookupById(stack, jsii.String("SG"), &AppConfig1.SecurityGroupID)

	// Get VPC
	PartVpc := awsec2.Vpc_FromLookup(stack, &AppConfig1.VPCid, &awsec2.VpcLookupOptions{VpcId: &AppConfig1.VPCid})

	// Find public subnets in the VPC
	publicSubnets := []awsec2.ISubnet{}

	// Iterate through all subnets in the VPC
	for _, subnet := range *PartVpc.PublicSubnets() {
		//if hasInternetRoute(subnet) {
		publicSubnets = append(publicSubnets, subnet)
		//}
	}

	// Create an IAM policy for EFS access
	RSarnEFS := "arn:aws:elasticfilesystem:*:" + AppConfig1.Account + ":file-system/*"
	efsPolicy := awsiam.NewPolicy(stack, jsii.String("EfsPolicy"), &awsiam.PolicyProps{
		PolicyName: jsii.String("EFSAccessPolicy"),
		Statements: &[]awsiam.PolicyStatement{
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("elasticfilesystem:ClientMount"),
					jsii.String("elasticfilesystem:ClientWrite"),
				},
				Resources: &[]*string{
					jsii.String(RSarnEFS),
				},
			}),
		},
	})
	// Create or reference the existing IAM role
	existingRoleArn := "arn:aws:iam::" + AppConfig1.Account + ":role/ecsTaskExecutionRole"
	existingRole := awsiam.Role_FromRoleArn(stack, jsii.String("ExistingRole"), &existingRoleArn, nil)
	efsPolicy.AttachToRole(existingRole)

	os.Setenv("AWS_SDK_LOAD_CONFIG", "true")

	sess, err := session.NewSession(&awsv1.Config{
		Region: aws.String(AppConfig1.Region),
	})

	if err != nil {
		log.Fatalf("❌ Error creating AWS session: %v", err)
		os.Exit(1)
	}

	// Create an AWS Secrets Manager service client
	svc := secretsmanagerv1.New(sess)
	SecretName := AppConfig.SecretNameSonarqube + AppConfig1.Index

	// Retrieve the existing secret value
	getSecretValueOutput, err := svc.GetSecretValue(&secretsmanagerv1.GetSecretValueInput{
		SecretId: &SecretName,
	})

	if err != nil {
		fmt.Println("❌ Error getting secret value:", err)
		os.Exit(1)
	}

	// Parse the existing secret JSON
	existingSecretJSON := *getSecretValueOutput.SecretString
	var secretStruct SecretStruct
	err = json.Unmarshal([]byte(existingSecretJSON), &secretStruct)
	if err != nil {
		fmt.Println("❌ Error parsing existing secret JSON:", err)
		os.Exit(1)
	}

	/*------------------------ Connect to RDS instance and create Database --------------------------*/

	// Set User/Password and JDBC URL for SonarQube Connexion
	sonarjdbc := AppConfig.Urlconnect + AppConfig.Endpoint + ":5432/sonarqube_part" + AppConfig1.Index + "?currentSchema=public"
	sonaruser := fmt.Sprintf("sonarqube_%s", AppConfig1.Index)
	sonarpass := secretStruct.DbPasswordPart

	/*--------------------------Set Inbount Rules------------------------------------------------------*/

	// Add Rules Inbount port 9000 on http
	sonarSG.AddIngressRule(awsec2.Peer_Ipv4(aws.String("0.0.0.0/0")),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			StringRepresentation: aws.String("Incoming web"),
			FromPort:             aws.Float64(9000),
			ToPort:               aws.Float64(9000),
		}),
		aws.String("Incoming http sonarQube"),
		aws.Bool(false),
	)
	// Add Rules Inbount port 2049 on NFS
	sonarSG.AddIngressRule(awsec2.Peer_Ipv4(aws.String("0.0.0.0/0")),
		awsec2.NewPort(&awsec2.PortProps{
			Protocol:             awsec2.Protocol_TCP,
			StringRepresentation: aws.String("Incoming efs"),
			FromPort:             aws.Float64(2049),
			ToPort:               aws.Float64(2049),
		}),
		aws.String("Incoming EFS Volume"),
		aws.Bool(false),
	)

	/*-----------------------------Created a PVC Volumes : EFS Data and Logs ----------------------*/

	// Create EFS Volume Data
	FSnamedata := AppConfig.SonarvolumeName1 + AppConfig1.Index
	FS := awsefs.NewFileSystem(stack, &AppConfig.SonarvolumeName1, &awsefs.FileSystemProps{
		FileSystemName:  &FSnamedata,
		Vpc:             PartVpc,
		PerformanceMode: awsefs.PerformanceMode_GENERAL_PURPOSE,
		RemovalPolicy:   awscdk.RemovalPolicy_DESTROY,
		SecurityGroup:   sonarSG,
		Encrypted:       aws.Bool(true),
		VpcSubnets: &awsec2.SubnetSelection{
			OnePerAz: aws.Bool(true),
		},
		ThroughputMode: awsefs.ThroughputMode_ELASTIC,
	})

	// Create an EFS Access Point Data
	accessPoint := FS.AddAccessPoint(jsii.String("DataAccessPoint1"), &awsefs.AccessPointOptions{
		Path: jsii.String("/opt/sonarqube/data"),
		CreateAcl: &awsefs.Acl{
			OwnerUid:    jsii.String("1000"),
			OwnerGid:    jsii.String("1000"),
			Permissions: jsii.String("0777"),
		},
		PosixUser: &awsefs.PosixUser{
			Uid: jsii.String("1000"),
			Gid: jsii.String("1000"),
		},
	})

	// Create EFS Volume Logs
	FSnamelog := AppConfig.SonarvolumeName2 + AppConfig1.Index
	FS2 := awsefs.NewFileSystem(stack, &AppConfig.SonarvolumeName2, &awsefs.FileSystemProps{
		FileSystemName:  &FSnamelog,
		Vpc:             PartVpc,
		PerformanceMode: awsefs.PerformanceMode_GENERAL_PURPOSE,
		RemovalPolicy:   awscdk.RemovalPolicy_DESTROY,
		SecurityGroup:   sonarSG,
		Encrypted:       aws.Bool(true),
		VpcSubnets: &awsec2.SubnetSelection{
			OnePerAz: aws.Bool(true),
		},
		ThroughputMode: awsefs.ThroughputMode_ELASTIC,
	})

	// Create an EFS Access Point Logs
	accessPoint2 := FS2.AddAccessPoint(jsii.String("LogsAccessPoint1"), &awsefs.AccessPointOptions{
		Path: jsii.String("/opt/sonarqube/logs"),
		CreateAcl: &awsefs.Acl{
			OwnerUid:    jsii.String("1000"),
			OwnerGid:    jsii.String("1000"),
			Permissions: jsii.String("0777"),
		},
		PosixUser: &awsefs.PosixUser{
			Uid: jsii.String("1000"),
			Gid: jsii.String("1000"),
		},
	})

	//-----------------------------------Set SonarQube Task--------------------------------------*/

	//Get a ECS cluster
	ClusterName := AppConfig.ClusterName + AppConfig1.Index
	ARNClust := "arn:aws:ecs:" + AppConfig1.Region + ":" + AppConfig1.Account + ":cluster/" + ClusterName
	securityGroup := awsec2.SecurityGroup_FromSecurityGroupId(stack, jsii.String("SecurityGroup"), jsii.String(AppConfig1.SecurityGroupID), nil)

	securityGroups := []awsec2.ISecurityGroup{securityGroup}
	serviceName := fmt.Sprintf("SVCPartner_%s", AppConfig1.Index)

	cluster := awsecs.Cluster_FromClusterAttributes(stack, aws.String("SonarFargateCluster"), &awsecs.ClusterAttributes{
		ClusterName:    &ClusterName,
		ClusterArn:     &ARNClust,
		Vpc:            PartVpc,
		SecurityGroups: &securityGroups,
	})

	// Set Ulimit parameter for SonarQube
	Ulimit := []*awsecs.Ulimit{
		{
			Name:      awsecs.UlimitName_NOFILE, // Ulimit name
			SoftLimit: jsii.Number(131072),      // Soft limit value
			HardLimit: jsii.Number(131072),      // Hard limit value
		},
		{
			Name:      awsecs.UlimitName_NPROC, // Ulimit name
			SoftLimit: jsii.Number(8192),       // Soft limit value
			HardLimit: jsii.Number(8192),       // Hard limit value
		},
	}

	// Create Fargate Task Definition & EfsVolumeConfiguration
	Taskname1 := AppConfig.Taskname + AppConfig1.Index
	taskDefinition := awsecs.NewFargateTaskDefinition(stack, &Taskname1, &awsecs.FargateTaskDefinitionProps{
		MemoryLimitMiB: jsii.Number(AppConfig.MemoryLimitMiB),
		Cpu:            jsii.Number(AppConfig.Cpu),
		Family:         jsii.String("SonarAppPart"),
		TaskRole:       existingRole,
		ExecutionRole:  existingRole,
	})

	// Set EFS Volume Data configuration
	efsVolumeConfiguration := &awsecs.EfsVolumeConfiguration{
		FileSystemId:      FS.FileSystemId(),
		TransitEncryption: jsii.String("ENABLED"),
		AuthorizationConfig: &awsecs.AuthorizationConfig{
			AccessPointId: accessPoint.AccessPointId(),
			Iam:           jsii.String("ENABLED"),
		},
	}
	// Set EFS Volume Logs configuration
	efsVolumeConfiguration2 := &awsecs.EfsVolumeConfiguration{
		FileSystemId:      FS2.FileSystemId(),
		TransitEncryption: jsii.String("ENABLED"),
		AuthorizationConfig: &awsecs.AuthorizationConfig{
			AccessPointId: accessPoint2.AccessPointId(),
			Iam:           jsii.String("ENABLED"),
		},
	}

	// Create EFS Volume Data
	volume := &awsecs.Volume{
		Name:                   &AppConfig.SonarvolumeName1,
		EfsVolumeConfiguration: efsVolumeConfiguration,
	}
	taskDefinition.AddVolume(volume)

	// Create EFS Volume Logs
	volume2 := &awsecs.Volume{
		Name:                   &AppConfig.SonarvolumeName2,
		EfsVolumeConfiguration: efsVolumeConfiguration2,
	}
	taskDefinition.AddVolume(volume2)

	// Set mount point Data
	mountPoint := &awsecs.MountPoint{
		ContainerPath: jsii.String("/opt/sonarqube/data"),
		ReadOnly:      aws.Bool(false),
		SourceVolume:  &AppConfig.SonarvolumeName1,
	}

	// Set mount point Logs
	mountPoint2 := &awsecs.MountPoint{
		ContainerPath: jsii.String("/opt/sonarqube/logs"),
		ReadOnly:      aws.Bool(false),
		SourceVolume:  &AppConfig.SonarvolumeName2,
	}

	// Set Container Options
	specificContainer := taskDefinition.AddContainer(jsii.String("SonarQubeContainer"), &awsecs.ContainerDefinitionOptions{
		Image:     awsecs.ContainerImage_FromRegistry(&AppConfig.SonarImages, &awsecs.RepositoryImageProps{}),
		Essential: aws.Bool(true),
		Command:   &[]*string{jsii.String("-Dsonar.search.javaAdditionalOpts=-Dnode.store.allow_mmap=false")},
		Environment: &map[string]*string{
			"SONAR_JDBC_URL":      &sonarjdbc,
			"SONAR_JDBC_USERNAME": &sonaruser,
			"SONAR_JDBC_PASSWORD": &sonarpass,
		},
		Logging: awsecs.LogDriver_AwsLogs(&awsecs.AwsLogDriverProps{
			StreamPrefix: jsii.String("Sonarqube"),
		}),
	})

	// Attach mount point Data in container
	specificContainer.AddMountPoints(mountPoint)
	// Attach mount point Logs in container
	specificContainer.AddMountPoints(mountPoint2)

	// Add Ulimits
	specificContainer.AddUlimits(Ulimit[0])
	specificContainer.AddUlimits(Ulimit[1])

	// Add Port mapping
	specificContainer.AddPortMappings(&awsecs.PortMapping{
		ContainerPort: jsii.Number(9000),
		HostPort:      jsii.Number(9000),
		Protocol:      awsecs.Protocol_TCP,
	})

	FgServices := awsecs.NewFargateService(stack, &serviceName, &awsecs.FargateServiceProps{
		Cluster:        cluster,
		TaskDefinition: taskDefinition,
		SecurityGroups: &[]awsec2.ISecurityGroup{securityGroup},
		AssignPublicIp: aws.Bool(true),
		VpcSubnets: &awsec2.SubnetSelection{
			OnePerAz: aws.Bool(true),
		},
		ServiceName: jsii.String(serviceName),
	})

	FgServices.Node().AddDependency(FS, accessPoint, FS2, accessPoint2)

	return stack
}

func createECSClient() *ecs.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load AWS SDK config")
	}

	return ecs.NewFromConfig(cfg)
}

func createEC2Client() *ec2.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load AWS SDK config")
	}

	return ec2.NewFromConfig(cfg)
}

func getTaskPublicIPAddress(clusterName, serviceName string) (string, error) {
	ecsClient := createECSClient()

	listTasksInput := &ecs.ListTasksInput{
		Cluster:     aws.String(clusterName),
		ServiceName: aws.String(serviceName),
	}

	listTasksOutput, err := ecsClient.ListTasks(context.TODO(), listTasksInput)
	if err != nil {
		return "", err
	}

	if len(listTasksOutput.TaskArns) == 0 {
		return "", fmt.Errorf("no tasks found for service %s in cluster %s", serviceName, clusterName)
	}

	taskArn := listTasksOutput.TaskArns[0]

	describeTasksInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(clusterName),
		Tasks:   []string{taskArn},
	}

	describeTasksOutput, err := ecsClient.DescribeTasks(context.TODO(), describeTasksInput)
	if err != nil {
		return "", err
	}

	if len(describeTasksOutput.Tasks) == 0 {
		return "", fmt.Errorf("no information found for task %s in cluster %s", taskArn, clusterName)
	}

	task := describeTasksOutput.Tasks[0]

	// Attempt to get public IP from network interface details
	if len(task.Attachments) > 0 {
		for _, attachment := range task.Attachments {
			for _, detail := range attachment.Details {
				if *detail.Name == "networkInterfaceId" && *detail.Value != "" {
					publicIP, err := getPublicIPFromNetworkInterface(*detail.Value)
					if err != nil {
						return "", err
					}
					return publicIP, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no public IP address found for task %s in cluster %s", taskArn, clusterName)
}

func getPublicIPFromNetworkInterface(networkInterfaceID string) (string, error) {
	ec2Client := createEC2Client()

	describeNetworkInterfacesInput := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{networkInterfaceID},
	}

	describeNetworkInterfacesOutput, err := ec2Client.DescribeNetworkInterfaces(context.TODO(), describeNetworkInterfacesInput)
	if err != nil {
		return "", err
	}

	if len(describeNetworkInterfacesOutput.NetworkInterfaces) == 0 {
		return "", fmt.Errorf("no information found for network interface %s", networkInterfaceID)
	}

	networkInterface := describeNetworkInterfacesOutput.NetworkInterfaces[0]

	for _, privateIPAddress := range networkInterface.PrivateIpAddresses {
		if privateIPAddress.Association != nil && privateIPAddress.Association.PublicIp != nil {
			return *privateIPAddress.Association.PublicIp, nil
		}
	}

	return "", fmt.Errorf("no public IP address found for network interface %s", networkInterfaceID)
}
func NewGetIpStack(scope constructs.Construct, id string, props *GetipStackProps, AppConfig Configuration, AppConfig1 ConfAuth) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	serviceName := fmt.Sprintf("SVCPartner_%s", AppConfig1.Index)
	Cluster := AppConfig.ClusterName + AppConfig1.Index
	//Region := AppConfig1.Region

	publicIPAddress, err := getTaskPublicIPAddress(Cluster, serviceName)
	if err != nil {
		fmt.Println("❌ Error get publicIPAddress:", err)
		os.Exit(1)
	}

	EndPoint := "http://" + publicIPAddress + ":9000"

	awscdk.NewCfnOutput(stack, jsii.String("SonarQube EndPoint "), &awscdk.CfnOutputProps{
		Value: &EndPoint,
	})

	return stack
}

func main() {

	var configcrd ConfAuth
	var config1 Configuration
	var AppConfig1, AppConfig = GetConfig(configcrd, config1)
	Stack := "SonarqubeStack" + AppConfig1.Index
	Stack2 := "GetPublicIP" + AppConfig1.Index

	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewSonarqubeStack(app, Stack, &SonarqubeStackProps{
		awscdk.StackProps{
			Env: env(AppConfig1.Region, AppConfig1.Account),
		},
	}, AppConfig, AppConfig1)

	NewGetIpStack(app, Stack2, &GetipStackProps{
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
