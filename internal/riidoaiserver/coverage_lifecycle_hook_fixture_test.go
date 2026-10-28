package riidoaiserver

func (f *coverageLifecycleDynamo) runPolicyFenceHook(target string, req coverageLifecycleRequest) {
	var hook func()
	if target == dynamoDBTransactWriteTarget && len(req.TransactItems) == 2 {
		policyFence := false
		conditionOnly := true
		for _, action := range req.TransactItems {
			if action.ConditionCheck == nil || action.Put != nil || action.Delete != nil || action.Update != nil {
				conditionOnly = false
				break
			}
			if action.ConditionCheck.ConditionExpression == executionCoverageCAS {
				policyFence = true
			}
		}
		if conditionOnly && policyFence {
			f.mu.Lock()
			hook = f.beforePolicyFence
			f.beforePolicyFence = nil
			f.mu.Unlock()
		}
	}
	if hook != nil {
		hook()
	}
}
