package consts

// P2P protocol IDs
const (
	// DeployProtocolID is the protocol ID for application deployment
	DeployProtocolID = "/p2p-playground/deploy/1.0.0"

	// ListProtocolID is the protocol ID for listing applications
	ListProtocolID = "/p2p-playground/list/1.0.0"

	// LogsProtocolID is the protocol ID for fetching application logs
	LogsProtocolID = "/p2p-playground/logs/1.0.0"
)

// System service constants
const (
	// DaemonServiceName is the name of the system service
	DaemonServiceName = "p2p-playground-daemon"
	// DaemonServiceDescription is the description of the system service
	DaemonServiceDescription = "P2P Playground Daemon - distributed application deployment node"
)
