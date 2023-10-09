package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	//"github.com/aws/aws-cdk-go/awscdk/awsservicediscovery"

	"github.com/aws/aws-cdk-go/awscdk/v2"

	//"github.com/aws/aws-cdk-lib/aws_route53"

	//"github.com/aws/aws-cdk-lib/aws_route53"

	//"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"

	//"github.com/aws/aws-cdk-lib/aws_route53"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsefs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"

	//"github.com/aws/aws-cdk-lib/aws_elasticloadbalancingv2"
	//"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	//"github.com/aws/aws-cdk-lib/aws_servicediscovery"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	//	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"

	//"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type SonarqubeStackProps struct {
	awscdk.StackProps
}

// Declare a struct for Config fields
type Configuration struct {
	SecretNameSonarqube string
	Region              string
	Endpoint            string
	SecretRDS           string
	PortDB              string
	MasterDB            string
	MasterUser          string
	Index               string
	SonarvolumeName1    string
	SonarvolumeName2    string
	ClusterName         string
	ClusterARN          string
	VPCid               string
	SecurityGroupID     string
	SonarImages         string
	Urlconnect          string
	UriDatabaseType     string
	DesiredCount        float64
	Cpu                 float64
	MemoryLimitMiB      float64
	Taskname            string
	EcsRole             string
}

//type SecretData struct {
//	DBPasswordpart string `json:"dbPasswordpart"`
//}

type RegionA struct {
	StringRegion []*string
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

func NewSonarqubeStack(scope constructs.Construct, id string, props *SonarqubeStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	var config1 Configuration
	var AppConfig = GetConfig(config1)

	// Get Security Group
	sonarSG := awsec2.SecurityGroup_FromLookupById(stack, jsii.String("SG"), &AppConfig.SecurityGroupID)
	//sonarSG := AppConfig.SecurityGroupID
	// Get VPC
	PartVpc := awsec2.Vpc_FromLookup(stack, &AppConfig.VPCid, &awsec2.VpcLookupOptions{VpcId: &AppConfig.VPCid})

	// Find public subnets in the VPC
	publicSubnets := []awsec2.ISubnet{}

	// Iterate through all subnets in the VPC
	for _, subnet := range *PartVpc.PublicSubnets() {
		// Check if the subnet has a route to the internet (indicating it's a public subnet)
		//if hasInternetRoute(subnet) {
		publicSubnets = append(publicSubnets, subnet)
		//}
	}

	// Create an IAM policy for EFS access
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
					jsii.String("arn:aws:elasticfilesystem:*:103078382956:file-system/*"),
					//jsii.String(RessourcesARNMT),
				},
			}),
		},
	})
	// Create or reference the existing IAM role
	existingRoleArn := AppConfig.EcsRole
	existingRole := awsiam.Role_FromRoleArn(stack, jsii.String("ExistingRole"), &existingRoleArn, nil)
	efsPolicy.AttachToRole(existingRole)

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(AppConfig.Region))
	if err != nil {
		log.Fatal(err)
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(AppConfig.SecretNameSonarqube),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString

	//var secretData map[string]interface{}
	//json.Unmarshal([]byte(secretString), &secretData)

	/*------------------------ Connect to RDS instance and create Database --------------------------*/

	// Set User/Password and JDBC URL for SonarQube Connexion
	sonarjdbc := AppConfig.Urlconnect + AppConfig.Endpoint + "/sonarqube_part" + AppConfig.Index
	sonaruser := fmt.Sprintf("sonarqube_%s", AppConfig.Index)
	sonarpass := secretString

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
	FSnamedata := AppConfig.SonarvolumeName1 + AppConfig.Index
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
	FSnamelog := AppConfig.SonarvolumeName2 + AppConfig.Index
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
	securityGroup := awsec2.SecurityGroup_FromSecurityGroupId(stack, jsii.String("SecurityGroup"), jsii.String(AppConfig.SecurityGroupID), nil)

	securityGroups := []awsec2.ISecurityGroup{securityGroup}
	serviceName := fmt.Sprintf("SVCPartner_%s", AppConfig.Index)

	cluster := awsecs.Cluster_FromClusterAttributes(stack, aws.String("SonarFargateCluster"), &awsecs.ClusterAttributes{
		ClusterName:    &AppConfig.ClusterName,
		ClusterArn:     &AppConfig.ClusterARN,
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
	Taskname1 := AppConfig.Taskname + AppConfig.Index
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

func main() {

	var config2 Configuration
	var AppConfig2 = GetConfig(config2)
	Stack := "SonarqubeStack" + AppConfig2.Index

	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewSonarqubeStack(app, Stack, &SonarqubeStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)

}

func env() *awscdk.Environment {

	return &awscdk.Environment{
		Account: jsii.String("xxxxxxxx"),
		Region:  jsii.String("eu-central-1"),
	}

}
