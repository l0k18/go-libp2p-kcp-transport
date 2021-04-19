package libp2pquic

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/libp2p/go-libp2p-quic-transport/metrics"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/lucas-clemente/quic-go/logging"
)

type statsConnectionTracer struct {
	qlogPath string
	metrics.ConnectionStats
}

func newStatsConnectionTracer(ctx context.Context, pers logging.Perspective, odcid logging.ConnectionID, node peer.ID, qlogPath string) *statsConnectionTracer {
	var remotePeer peer.ID
	if v := ctx.Value(dialTracingKey); v != nil {
		remotePeer = v.(peer.ID)
	}
	t := &statsConnectionTracer{qlogPath: qlogPath}
	t.ConnectionStats.ODCID = odcid
	t.ConnectionStats.Node = node
	t.ConnectionStats.Perspective = pers
	t.ConnectionStats.Peer = remotePeer
	return t
}

func (t *statsConnectionTracer) StartedConnection(local, remote net.Addr, srcConnID, destConnID logging.ConnectionID) {
	t.ConnectionStats.StartTime = time.Now()
	t.ConnectionStats.LocalAddr = local
	t.ConnectionStats.RemoteAddr = remote
}

func (t *statsConnectionTracer) ClosedConnection(r logging.CloseReason) {
	t.ConnectionStats.CloseReason = r
	t.ConnectionStats.EndTime = time.Now()
}
func (t *statsConnectionTracer) SentTransportParameters(*logging.TransportParameters) {}
func (t *statsConnectionTracer) ReceivedTransportParameters(tp *logging.TransportParameters) {
	if t.Perspective == logging.PerspectiveServer {
		if tp.RetrySourceConnectionID != nil {
			t.ConnectionStats.PacketsSent.Retry++
		}
	}
}
func (t *statsConnectionTracer) RestoredTransportParameters(*logging.TransportParameters) {}

func (t *statsConnectionTracer) countPacket(c *metrics.PacketCounter, hdr *logging.ExtendedHeader) {
	switch logging.PacketTypeFromHeader(&hdr.Header) {
	case logging.PacketTypeInitial:
		c.Initial++
	case logging.PacketTypeHandshake:
		c.Handshake++
	case logging.PacketType0RTT:
		c.ZeroRTT++
	case logging.PacketTypeRetry:
		c.Retry++
	case logging.PacketType1RTT:
		c.ShortHeader++
	}
}

func (t *statsConnectionTracer) SentPacket(hdr *logging.ExtendedHeader, _ logging.ByteCount, _ *logging.AckFrame, _ []logging.Frame) {
	t.countPacket(&t.ConnectionStats.PacketsSent, hdr)
}

func (t *statsConnectionTracer) ReceivedVersionNegotiationPacket(_ *logging.Header, v []logging.VersionNumber) {
	t.ConnectionStats.VersionNegotiationVersions = v
}

func (t *statsConnectionTracer) ReceivedRetry(*logging.Header) {
	t.ConnectionStats.PacketsRcvd.Retry++
}

func (t *statsConnectionTracer) ReceivedPacket(hdr *logging.ExtendedHeader, _ logging.ByteCount, _ []logging.Frame) {
	t.countPacket(&t.ConnectionStats.PacketsRcvd, hdr)
}

func (t *statsConnectionTracer) BufferedPacket(logging.PacketType) {
	t.ConnectionStats.PacketsBuffered++
}

func (t *statsConnectionTracer) DroppedPacket(logging.PacketType, logging.ByteCount, logging.PacketDropReason) {
	t.ConnectionStats.PacketsDropped++
}

func (t *statsConnectionTracer) UpdatedMetrics(rttStats *logging.RTTStats, cwnd, bytesInFlight logging.ByteCount, packetsInFlight int) {
	t.ConnectionStats.LastRTT = metrics.RTTMeasurement{
		SmoothedRTT: rttStats.SmoothedRTT(),
		RTTVar:      rttStats.MeanDeviation(),
		MinRTT:      rttStats.MinRTT(),
	}
}

func (t *statsConnectionTracer) LostPacket(encLevel logging.EncryptionLevel, _ logging.PacketNumber, _ logging.PacketLossReason) {
	switch encLevel {
	case logging.EncryptionInitial:
		t.ConnectionStats.PacketsLost.Initial++
	case logging.EncryptionHandshake:
		t.ConnectionStats.PacketsLost.Handshake++
	case logging.Encryption0RTT:
		t.ConnectionStats.PacketsLost.ZeroRTT++
	case logging.Encryption1RTT:
		t.ConnectionStats.PacketsLost.ShortHeader++
	}
}

func (t *statsConnectionTracer) AcknowledgedPacket(logging.EncryptionLevel, logging.PacketNumber) {}

func (t *statsConnectionTracer) UpdatedCongestionState(logging.CongestionState) {}
func (t *statsConnectionTracer) UpdatedPTOCount(value uint32) {
	if value > 0 {
		t.ConnectionStats.PTOCount++
	}
}

func (t *statsConnectionTracer) UpdatedKeyFromTLS(l logging.EncryptionLevel, p logging.Perspective) {
	if l == logging.Encryption1RTT && p == logging.PerspectiveClient {
		t.ConnectionStats.HandshakeCompleteTime = time.Now()
		t.ConnectionStats.HandshakeRTT = t.ConnectionStats.LastRTT
	}
}
func (t *statsConnectionTracer) UpdatedKey(generation logging.KeyPhase, remote bool)                {}
func (t *statsConnectionTracer) DroppedEncryptionLevel(logging.EncryptionLevel)                     {}
func (t *statsConnectionTracer) DroppedKey(generation logging.KeyPhase)                             {}
func (t *statsConnectionTracer) SetLossTimer(logging.TimerType, logging.EncryptionLevel, time.Time) {}
func (t *statsConnectionTracer) LossTimerExpired(logging.TimerType, logging.EncryptionLevel)        {}
func (t *statsConnectionTracer) LossTimerCanceled()                                                 {}

// Close is called when the connection is closed.
func (t *statsConnectionTracer) Close() {
	if len(t.qlogPath) > 0 {
		qlog, err := t.saveQlog()
		if err != nil {
			log.Errorf("Error saving qlog: %s", err)
		} else {
			t.ConnectionStats.Qlog = qlog
		}
	}
	if t.ConnectionStats.StartTime.IsZero() { // Close() called before StartedConnection()
		return
	}
	if err := t.ConnectionStats.Save(); err != nil {
		log.Errorf("Saving connection statistics failed: %s", err)
	}
}

func (t *statsConnectionTracer) saveQlog() (string, error) {
	f, err := os.Open(t.qlogPath)
	if err != nil {
		return "", fmt.Errorf("failed to open qlog file %s: %w", t.qlogPath, err)
	}
	return metrics.Upload(filepath.Base(f.Name()), f)
}

func (t *statsConnectionTracer) Debug(name, msg string) {}

var _ logging.ConnectionTracer = &statsConnectionTracer{}
