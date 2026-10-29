package nex






type ConnectionState uint8

const (

	StateNotConnected ConnectionState = iota


	StateConnecting


	StateConnected


	StateDisconnecting


	StateFaulty
)
