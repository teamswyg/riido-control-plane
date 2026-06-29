package main

func verifyLoopCoverageDimensions(m manifest) error {
	for _, dim := range loopCoverageDimensions {
		if err := verifyLoopCoverageDimension(m, dim); err != nil {
			return err
		}
	}
	return nil
}
