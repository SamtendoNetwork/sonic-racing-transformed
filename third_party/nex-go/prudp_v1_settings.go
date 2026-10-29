package nex

import "net"




type PRUDPV1Settings struct {
	LegacyConnectionSignature     bool
	ConnectionSignatureCalculator func(packet *PRUDPPacketV1, addr net.Addr) ([]byte, error)
	SignatureCalculator           func(packet *PRUDPPacketV1, sessionKey, connectionSignature []byte) []byte
}


func NewPRUDPV1Settings() *PRUDPV1Settings {
	return &PRUDPV1Settings{
		LegacyConnectionSignature:     false,
		ConnectionSignatureCalculator: defaultPRUDPv1ConnectionSignature,
		SignatureCalculator:           defaultPRUDPv1CalculateSignature,
	}
}
