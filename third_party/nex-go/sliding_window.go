package nex






type SlidingWindow struct {
	sequenceIDCounter *Counter[uint16]
	streamSettings    *StreamSettings
	TimeoutManager    *TimeoutManager
}


func (sw *SlidingWindow) SetCipherKey(key []byte) {
	sw.streamSettings.EncryptionAlgorithm.SetKey(key)
}


func (sw *SlidingWindow) NextOutgoingSequenceID() uint16 {
	return sw.sequenceIDCounter.Next()
}


func (sw *SlidingWindow) Decrypt(data []byte) ([]byte, error) {
	return sw.streamSettings.EncryptionAlgorithm.Decrypt(data)
}


func (sw *SlidingWindow) Encrypt(data []byte) ([]byte, error) {
	return sw.streamSettings.EncryptionAlgorithm.Encrypt(data)
}


func NewSlidingWindow() *SlidingWindow {
	sw := &SlidingWindow{
		sequenceIDCounter: NewCounter[uint16](0),
		TimeoutManager:    NewTimeoutManager(),
	}

	return sw
}
