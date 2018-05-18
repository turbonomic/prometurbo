package version

// Remote Mediation Clients are expected to send the protobuf message version to ensure
// compatibility with the server.
// This version string is sent as part of the registration protocol in the Negotiation message.
// Client probes are allowed to register with the server only if the server accepts the version.
const (
	PROTOBUF_VERSION string = "6.1.0"
)
