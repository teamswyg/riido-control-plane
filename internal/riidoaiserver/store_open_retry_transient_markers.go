package riidoaiserver

var transientStoreOpenErrorMarkers = []string{
	"ThrottlingException",
	"ProvisionedThroughputExceededException",
	"RequestLimitExceeded",
	"TooManyRequestsException",
	"InternalServerError",
	"ServiceUnavailable",
}
