package awsadapters

import internal "github.com/teamswyg/riido-control-plane/internal/riidoaiserver"

const (
	SchemaVersion                         = internal.SchemaVersion
	StoreSnapshotSchemaVersion            = internal.StoreSnapshotSchemaVersion
	OutboxRecordSchemaVersion             = internal.OutboxRecordSchemaVersion
	AIAgentClientPersistenceSchemaVersion = internal.AIAgentClientPersistenceSchemaVersion

	AssignmentQueued      = internal.AssignmentQueued
	EventAssignmentQueued = internal.EventAssignmentQueued
)

type (
	AssignmentState = internal.AssignmentState
	Assignment      = internal.Assignment
	TaskEvent       = internal.TaskEvent
	OutboxRecord    = internal.OutboxRecord

	AWSCredentials                        = internal.AWSCredentials
	AWSCredentialsProvider                = internal.AWSCredentialsProvider
	StaticAWSCredentialsProvider          = internal.StaticAWSCredentialsProvider
	ECSContainerCredentialsProvider       = internal.ECSContainerCredentialsProvider
	ECSContainerCredentialsProviderConfig = internal.ECSContainerCredentialsProviderConfig

	DynamoDBOutboxConfig = internal.DynamoDBOutboxConfig
	DynamoDBOutbox       = internal.DynamoDBOutbox

	StoreSnapshot               = internal.StoreSnapshot
	StoreSnapshotTask           = internal.StoreSnapshotTask
	DynamoDBStoreSnapshotConfig = internal.DynamoDBStoreSnapshotConfig
	DynamoDBStoreSnapshot       = internal.DynamoDBStoreSnapshot

	AIAgentClientSnapshot                 = internal.AIAgentClientSnapshot
	AIAgentClientDeviceCredentialSnapshot = internal.AIAgentClientDeviceCredentialSnapshot
	AIAgentClientEventSnapshot            = internal.AIAgentClientEventSnapshot
	DynamoDBAIAgentClientSnapshotConfig   = internal.DynamoDBAIAgentClientSnapshotConfig
	DynamoDBAIAgentClientSnapshot         = internal.DynamoDBAIAgentClientSnapshot

	AssignmentOperationRecord              = internal.AssignmentOperationRecord
	AssignmentClaimResult                  = internal.AssignmentClaimResult
	AssignmentActiveLease                  = internal.AssignmentActiveLease
	AssignmentProjection                   = internal.AssignmentProjection
	DynamoDBAssignmentOperationStoreConfig = internal.DynamoDBAssignmentOperationStoreConfig
	DynamoDBAssignmentOperationStore       = internal.DynamoDBAssignmentOperationStore

	DynamoDBTableStreamARNConfig = internal.DynamoDBTableStreamARNConfig
	DynamoDBTableDescription     = internal.DynamoDBTableDescription
	DynamoDBKeySchemaMember      = internal.DynamoDBKeySchemaMember

	StreamRelayPublisher       = internal.StreamRelayPublisher
	StreamRelayCheckpointStore = internal.StreamRelayCheckpointStore
	StreamRelayEvent           = internal.StreamRelayEvent
	StreamRelayCheckpoint      = internal.StreamRelayCheckpoint
	DynamoDBStreamRelayConfig  = internal.DynamoDBStreamRelayConfig
	DynamoDBStreamRelay        = internal.DynamoDBStreamRelay
	DynamoDBStreamRelayStats   = internal.DynamoDBStreamRelayStats

	DynamoDBStreamRelayCheckpointStoreConfig = internal.DynamoDBStreamRelayCheckpointStoreConfig
	DynamoDBStreamRelayCheckpointStore       = internal.DynamoDBStreamRelayCheckpointStore

	EventBridgeStreamRelayPublisherConfig = internal.EventBridgeStreamRelayPublisherConfig
	EventBridgeStreamRelayPublisher       = internal.EventBridgeStreamRelayPublisher
)

var (
	NewStaticAWSCredentialsProvider    = internal.NewStaticAWSCredentialsProvider
	NewECSContainerCredentialsProvider = internal.NewECSContainerCredentialsProvider

	NewDynamoDBOutbox                   = internal.NewDynamoDBOutbox
	NewDynamoDBStoreSnapshot            = internal.NewDynamoDBStoreSnapshot
	NewDynamoDBAIAgentClientSnapshot    = internal.NewDynamoDBAIAgentClientSnapshot
	NewDynamoDBAssignmentOperationStore = internal.NewDynamoDBAssignmentOperationStore

	LoadDynamoDBTableStreamARN = internal.LoadDynamoDBTableStreamARN
	DescribeDynamoDBTable      = internal.DescribeDynamoDBTable

	NewDynamoDBStreamRelay                = internal.NewDynamoDBStreamRelay
	RunDynamoDBStreamRelayOnce            = internal.RunDynamoDBStreamRelayOnce
	NewDynamoDBStreamRelayCheckpointStore = internal.NewDynamoDBStreamRelayCheckpointStore

	NewEventBridgeStreamRelayPublisher = internal.NewEventBridgeStreamRelayPublisher
)
