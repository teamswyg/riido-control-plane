package riidoaiserver

import (
	"context"
	"math"
	"strconv"
)

func bootstrapFenceItem(r deviceBootstrapRecord) map[string]map[string]string {
	item := devicePolicyKey(r.Before.DeviceID)
	item["policy_schema"] = map[string]string{"S": deviceBootstrapFenceSchema}
	item["coverage_state"] = map[string]string{"S": "bootstrap_pending"}
	item["transition_id"] = map[string]string{"S": r.Request.TransitionID}
	item["revision"] = map[string]string{"N": strconv.FormatUint(r.PolicyRevision, 10)}
	return item
}

func bootstrapFenceValues(r deviceBootstrapRecord) map[string]map[string]string {
	return map[string]map[string]string{":schema": {"S": deviceBootstrapFenceSchema}, ":state": {"S": "bootstrap_pending"}, ":transition": {"S": r.Request.TransitionID}, ":revision": {"N": strconv.FormatUint(r.PolicyRevision, 10)}}
}

const bootstrapFenceCondition = "policy_schema = :schema AND coverage_state = :state AND transition_id = :transition AND revision = :revision"

func (s *DynamoDBAIAgentClientSnapshot) bootstrapPolicyAction(ctx context.Context, mutation deviceBootstrapMutation, credentials AWSCredentials) (dynamoDBTransactWriteItem, error) {
	r, before := mutation.after, mutation.before
	if r.State == "issued" {
		return legacyExecutionAdmissionFence(s.tableName, r.Before.DeviceID), nil
	}
	if before.State == "issued" {
		item, err := s.bootstrapGet(ctx, devicePolicyKey(r.Before.DeviceID), credentials)
		if err != nil {
			return dynamoDBTransactWriteItem{}, err
		}
		revision, err := devicePolicyRevision(item, r.Before.DeviceID)
		if err != nil || revision >= math.MaxUint64-1 || revision+1 != r.PolicyRevision {
			return dynamoDBTransactWriteItem{}, errCLIInstallationConflict
		}
		put := &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: bootstrapFenceItem(r), ConditionExpression: "attribute_not_exists(pk) AND attribute_not_exists(sk)"}
		if len(item) != 0 {
			put.ConditionExpression = "policy_schema = :schema AND coverage_state = :state AND revision = :revision"
			put.ExpressionAttributeValues = map[string]map[string]string{":schema": {"S": devicePolicyRevisionSchema}, ":state": {"S": "unknown"}, ":revision": {"N": strconv.FormatUint(revision, 10)}}
		}
		return dynamoDBTransactWriteItem{Put: put}, nil
	}
	if r.State == "prepared" {
		return dynamoDBTransactWriteItem{ConditionCheck: &dynamoDBTransactWriteConditionCheckAction{TableName: s.tableName, Key: devicePolicyKey(r.Before.DeviceID), ConditionExpression: bootstrapFenceCondition, ExpressionAttributeValues: bootstrapFenceValues(*before)}}, nil
	}
	item, err := executionCoverageItem(executionCoverage{DeviceID: r.After.DeviceID, Generation: r.After.ExecutionCoverageGeneration, CredentialGeneration: r.After.CredentialGeneration, OwnerPrincipalID: r.After.OwnerPrincipalID, State: "complete", Revision: r.PolicyRevision})
	if err != nil {
		return dynamoDBTransactWriteItem{}, err
	}
	return dynamoDBTransactWriteItem{Put: &dynamoDBTransactWritePutAction{TableName: s.tableName, Item: item, ConditionExpression: bootstrapFenceCondition, ExpressionAttributeValues: bootstrapFenceValues(*before)}}, nil
}
