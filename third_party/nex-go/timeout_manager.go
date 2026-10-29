package nex

import (
	"context"
	"time"
)


type TimeoutManager struct {
	ctx            context.Context
	cancel         context.CancelFunc
	packets        *MutexMap[uint16, PRUDPPacketInterface]
	streamSettings *StreamSettings
}


func (tm *TimeoutManager) SchedulePacketTimeout(packet PRUDPPacketInterface) {
	endpoint := packet.Sender().Endpoint().(*PRUDPEndPoint)

	rto := endpoint.ComputeRetransmitTimeout(packet)
	ctx, cancel := context.WithTimeout(tm.ctx, rto)

	timeout := NewTimeout()
	timeout.SetRTO(rto)
	timeout.ctx = ctx
	timeout.cancel = cancel
	packet.setTimeout(timeout)

	tm.packets.Set(packet.SequenceID(), packet)
	go tm.start(packet)
}


func (tm *TimeoutManager) AcknowledgePacket(sequenceID uint16) {

	tm.packets.RunAndDelete(sequenceID, func(_ uint16, packet PRUDPPacketInterface) {

		if packet.SendCount() >= tm.streamSettings.RTTRetransmit {
			rttm := time.Since(packet.SentAt())
			packet.Sender().(*PRUDPConnection).rtt.Adjust(rttm)
		}
	})
}

func (tm *TimeoutManager) start(packet PRUDPPacketInterface) {
	<-packet.getTimeout().ctx.Done()

	connection := packet.Sender().(*PRUDPConnection)


	if connection.ConnectionState != StateConnected {
		return
	}

	if tm.packets.Has(packet.SequenceID()) {
		endpoint := packet.Sender().Endpoint().(*PRUDPEndPoint)


		if packet.SendCount() < tm.streamSettings.MaxPacketRetransmissions {
			packet.incrementSendCount()
			packet.setSentAt(time.Now())
			rto := endpoint.ComputeRetransmitTimeout(packet)

			ctx, cancel := context.WithTimeout(tm.ctx, rto)
			timeout := packet.getTimeout()
			timeout.timeout = rto
			timeout.ctx = ctx
			timeout.cancel = cancel


			go tm.start(packet)


			server := connection.endpoint.Server
			data := packet.Bytes()
			server.SendRaw(connection.Socket, data)
		} else {

			endpoint.CleanupConnection(connection)
		}
	}
}


func (tm *TimeoutManager) Stop() {
	tm.cancel()
	tm.packets.Clear(func(key uint16, value PRUDPPacketInterface) {})
}


func NewTimeoutManager() *TimeoutManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TimeoutManager{
		ctx:            ctx,
		cancel:         cancel,
		packets:        NewMutexMap[uint16, PRUDPPacketInterface](),
		streamSettings: NewStreamSettings(),
	}
}
