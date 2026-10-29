package nex

import (
	"github.com/PretendoNetwork/nex-go/v2/compression"
	"github.com/PretendoNetwork/nex-go/v2/encryption"
)







type StreamSettings struct {
	ExtraRetransmitTimeoutTrigger    uint32
	MaxPacketRetransmissions         uint32
	KeepAliveTimeout                 uint32
	ChecksumBase                     uint32
	FaultDetectionEnabled            bool
	InitialRTT                       uint32
	SynInitialRTT                    uint32
	EncryptionAlgorithm              encryption.Algorithm
	ExtraRetransmitTimeoutMultiplier float32
	WindowSize                       uint32
	CompressionAlgorithm             compression.Algorithm
	RTTRetransmit                    uint32
	RetransmitTimeoutMultiplier      float32
	MaxSilenceTime                   uint32
}


func (ss *StreamSettings) Copy() *StreamSettings {
	copied := NewStreamSettings()

	copied.ExtraRetransmitTimeoutTrigger = ss.ExtraRetransmitTimeoutTrigger
	copied.MaxPacketRetransmissions = ss.MaxPacketRetransmissions
	copied.KeepAliveTimeout = ss.KeepAliveTimeout
	copied.ChecksumBase = ss.ChecksumBase
	copied.FaultDetectionEnabled = ss.FaultDetectionEnabled
	copied.InitialRTT = ss.InitialRTT
	copied.EncryptionAlgorithm = ss.EncryptionAlgorithm.Copy()
	copied.ExtraRetransmitTimeoutMultiplier = ss.ExtraRetransmitTimeoutMultiplier
	copied.WindowSize = ss.WindowSize
	copied.CompressionAlgorithm = ss.CompressionAlgorithm.Copy()
	copied.RTTRetransmit = ss.RTTRetransmit
	copied.RetransmitTimeoutMultiplier = ss.RetransmitTimeoutMultiplier
	copied.MaxSilenceTime = ss.MaxSilenceTime

	return copied
}


func NewStreamSettings() *StreamSettings {


	return &StreamSettings{
		ExtraRetransmitTimeoutTrigger:    0x32,
		MaxPacketRetransmissions:         0x14,
		KeepAliveTimeout:                 1000,
		ChecksumBase:                     0,
		FaultDetectionEnabled:            true,
		InitialRTT:                       0x2EE,
		SynInitialRTT:                    0xFA,
		EncryptionAlgorithm:              encryption.NewRC4Encryption(),
		ExtraRetransmitTimeoutMultiplier: 1.0,
		WindowSize:                       8,
		CompressionAlgorithm:             compression.NewDummyCompression(),
		RTTRetransmit:                    2,
		RetransmitTimeoutMultiplier:      1.25,
		MaxSilenceTime:                   10000,
	}
}
