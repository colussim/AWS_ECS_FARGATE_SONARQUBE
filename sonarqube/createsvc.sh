aws ecs create-service \
  --cluster ClustPartner01 \
  --service-name sonartest\
  --task-definition test-task \
  --desired-count 1 \
  --launch-type EC2 \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-0f469ea6710ec7976,subnet-05885fc4f55a77b95],securityGroups=[sg-0d8ac80b189512b0f]}"

