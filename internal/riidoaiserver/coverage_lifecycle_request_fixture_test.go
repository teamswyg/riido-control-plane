package riidoaiserver

type coverageLifecycleRequest struct {
	coverageLifecycleAction
	IndexName, KeyConditionExpression string
	Limit                             int
	ExclusiveStartKey                 map[string]map[string]string
	TransactItems                     []struct {
		Put, Get, Delete, ConditionCheck *coverageLifecycleAction
		Update                           *dynamoDBTransactWriteUpdateAction
	}
}
