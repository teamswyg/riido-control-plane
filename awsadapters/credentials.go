package awsadapters

import internal "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

type (
	AWSCredentials                        = internal.AWSCredentials
	AWSCredentialsProvider                = internal.AWSCredentialsProvider
	StaticAWSCredentialsProvider          = internal.StaticAWSCredentialsProvider
	ECSContainerCredentialsProvider       = internal.ECSContainerCredentialsProvider
	ECSContainerCredentialsProviderConfig = internal.ECSContainerCredentialsProviderConfig
)

var (
	NewStaticAWSCredentialsProvider    = internal.NewStaticAWSCredentialsProvider
	NewECSContainerCredentialsProvider = internal.NewECSContainerCredentialsProvider
)
