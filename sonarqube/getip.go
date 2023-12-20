package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

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

func main() {

	clusterName := "ClustPartner02"
	serviceName := "SVCPartner_02"

	publicIPAddress, err := getTaskPublicIPAddress(clusterName, serviceName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Public IP address for task in cluster %s and service %s: %s\n", clusterName, serviceName, publicIPAddress)
}
