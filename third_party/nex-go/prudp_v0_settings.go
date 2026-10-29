package nex

import "net"




type PRUDPV0Settings struct {
	IsQuazalMode                  bool
	EncryptedConnect              bool
	LegacyConnectionSignature     bool
	UseEnhancedChecksum           bool
	ConnectionSignatureCalculator func(packet *PRUDPPacketV0, addr net.Addr) ([]byte, error)
	SignatureCalculator           func(packet *PRUDPPacketV0, sessionKey, connectionSignature []byte) []byte
	DataSignatureCalculator       func(packet *PRUDPPacketV0, sessionKey []byte) []byte
	ChecksumCalculator            func(packet *PRUDPPacketV0, data []byte) uint32
}


func NewPRUDPV0Settings() *PRUDPV0Settings {
	return &PRUDPV0Settings{
		IsQuazalMode:                  false,
		EncryptedConnect:              false,
		LegacyConnectionSignature:     false,
		UseEnhancedChecksum:           false,
		ConnectionSignatureCalculator: defaultPRUDPv0ConnectionSignature,
		SignatureCalculator:           defaultPRUDPv0CalculateSignature,
		DataSignatureCalculator:       defaultPRUDPv0CalculateDataSignature,
		ChecksumCalculator:            defaultPRUDPv0CalculateChecksum,
	}
}
