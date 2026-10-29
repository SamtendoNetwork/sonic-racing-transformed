package nex





type PacketDispatchQueue struct {
	queue                  map[uint16]PRUDPPacketInterface
	nextExpectedSequenceId *Counter[uint16]
}


func (pdq *PacketDispatchQueue) Queue(packet PRUDPPacketInterface) {
	pdq.queue[packet.SequenceID()] = packet
}



func (pdq *PacketDispatchQueue) GetNextToDispatch() (PRUDPPacketInterface, bool) {
	if packet, ok := pdq.queue[pdq.nextExpectedSequenceId.Value]; ok {
		return packet, true
	}

	return nil, false
}


func (pdq *PacketDispatchQueue) Dispatched(packet PRUDPPacketInterface) {
	pdq.nextExpectedSequenceId.Next()
	delete(pdq.queue, packet.SequenceID())
}


func (pdq *PacketDispatchQueue) Purge() {
	clear(pdq.queue)
}


func NewPacketDispatchQueue() *PacketDispatchQueue {
	return &PacketDispatchQueue{
		queue:                  make(map[uint16]PRUDPPacketInterface),
		nextExpectedSequenceId: NewCounter[uint16](2),
	}
}
