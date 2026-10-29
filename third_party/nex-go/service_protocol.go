package nex


type ServiceProtocol interface {
	HandlePacket(packet PacketInterface)
	Endpoint() EndpointInterface
	SetEndpoint(endpoint EndpointInterface)
}
