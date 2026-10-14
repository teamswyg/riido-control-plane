package awsadapters

import internal "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

type (
	DynamoDBTableStreamARNConfig = internal.DynamoDBTableStreamARNConfig
	DynamoDBTableDescription     = internal.DynamoDBTableDescription
	DynamoDBKeySchemaMember      = internal.DynamoDBKeySchemaMember
)

var (
	LoadDynamoDBTableStreamARN = internal.LoadDynamoDBTableStreamARN
	DescribeDynamoDBTable      = internal.DescribeDynamoDBTable
)
