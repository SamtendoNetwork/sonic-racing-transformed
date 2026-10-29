package nex


type EndpointInterface interface {
	AccessKey() string
	SetAccessKey(accessKey string)
	Send(packet PacketInterface)
	LibraryVersions() *LibraryVersions
	ByteStreamSettings() *ByteStreamSettings
	SetByteStreamSettings(settings *ByteStreamSettings)
	UseVerboseRMC() bool
	EnableVerboseRMC(enabled bool)
	EmitError(err *Error)
}
