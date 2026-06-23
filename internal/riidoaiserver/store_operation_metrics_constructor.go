package riidoaiserver

func NewStoreOperationMetrics() *StoreOperationMetrics {
	return &StoreOperationMetrics{
		buckets: map[int64]*storeOperationBucket{},
	}
}
