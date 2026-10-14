package riidoaiserver

func withStreamRelayEventSchema(event StreamRelayEvent, schema string) StreamRelayEvent {
	event.SchemaVersion = schema
	return event
}

func withOutboxRecordSchema(event StreamRelayEvent, schema string) StreamRelayEvent {
	event.Record.SchemaVersion = schema
	return event
}
