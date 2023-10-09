import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/aws/external"
    "github.com/aws/aws-sdk-go-v2/service/rds"
)

func checkRDSStatusAndRetrieveConfig(instanceIdentifier, region string) (*rds.DBInstance, error) {
    cfg, err := external.LoadDefaultAWSConfig()
    if err != nil {
        return nil, err
    }

    cfg.Region = region

    // Create an RDS client
    svc := rds.New(cfg)

    // Describe the RDS instance
    input := &rds.DescribeDBInstancesInput{
        DBInstanceIdentifier: aws.String(instanceIdentifier),
    }

    resp, err := svc.DescribeDBInstances(context.TODO(), input)
    if err != nil {
        return nil, err
    }

    if len(resp.DBInstances) == 0 {
        return nil, fmt.Errorf("RDS instance not found")
    }

    // Assuming you have only one instance with the given identifier
    return &resp.DBInstances[0], nil
}

func main() {
    instanceIdentifier := "your-db-instance-id" // Replace with your RDS instance ID
    region := "us-east-1"                      // Replace with your desired region

    dbInstance, err := checkRDSStatusAndRetrieveConfig(instanceIdentifier, region)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Printf("RDS Instance Status: %s\n", dbInstance.DBInstanceStatus)
    fmt.Printf("Endpoint: %s\n", *dbInstance.Endpoint.Address)
    fmt.Printf("Port: %d\n", *dbInstance.Endpoint.Port)
    fmt.Printf("Username: %s\n", *dbInstance.MasterUsername)
}

