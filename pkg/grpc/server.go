package grpc

import (
	"io"

	"github.com/google/uuid"
	"github.com/kubensage/relay/proto/gen"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MetricsServer implements the gRPC MetricsServiceServer interface.
//
// Responsibilities:
//   - Accepts streamed metrics from agents via SendMetrics.
//   - Fans out incoming metrics to all active subscribers via a Broadcaster.
//   - Allows clients to subscribe to a live metrics stream via SubscribeMetrics.
type MetricsServer struct {
	gen.UnimplementedMetricsServiceServer
	broadcaster *Broadcaster // Manages subscribers and broadcasts messages
	logger      *zap.Logger  // Structured logger for observability
}

// NewMetricsServer creates a new MetricsServer.
//
// Parameters:
//   - logger: zap.Logger for structured logging.
//
// Returns:
//   - *MetricsServer: initialized server ready to be registered with gRPC.
func NewMetricsServer(logger *zap.Logger) *MetricsServer {
	return &MetricsServer{
		broadcaster: NewBroadcaster(logger),
		logger:      logger,
	}
}

// SendMetrics handles incoming streamed metrics from agents.
//
// Behavior:
//   - Continuously reads from the gRPC stream until EOF or error.
//   - Each received message is logged at INFO level (host, pod count).
//   - Messages are broadcasted to all active subscribers.
//   - On EOF, an acknowledgment is returned to the agent.
//
// Parameters:
//   - stream: gRPC server stream used by agents to send Metrics messages.
//
// Returns:
//   - error: if reading from the stream fails or acknowledgment cannot be sent.
func (s *MetricsServer) SendMetrics(stream gen.MetricsService_SendMetricsServer) error {
	s.logger.Info("started receiving metrics from agent")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			s.logger.Info("agent stream closed, sending acknowledgment")
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			s.logger.Error("failed to receive metrics from agent", zap.Error(err))
			return err
		}

		s.logger.Info("received metrics batch",
			zap.String("host", req.GetNodeMetrics().GetHostname()),
			zap.Int("pods_count", len(req.GetPodMetrics())),
		)

		s.broadcaster.Broadcast(req)
	}
}

// SubscribeMetrics allows a client to subscribe to the live metrics stream.
//
// Behavior:
//   - Assigns a unique ID to the subscriber.
//   - Registers the subscriber with a buffered channel.
//   - Streams metrics to the client until the context is canceled or an error occurs.
//   - Ensures cleanup on disconnect.
//
// Parameters:
//   - _ (*emptypb.Empty): unused input placeholder.
//   - stream: gRPC stream used to send metrics messages to the subscriber.
//
// Returns:
//   - error: if sending fails or the stream context is canceled.
func (s *MetricsServer) SubscribeMetrics(_ *emptypb.Empty, stream gen.MetricsService_SubscribeMetricsServer) error {
	id := uuid.New().String()
	ch := make(chan *gen.Metrics, 100)

	s.logger.Info("subscriber connected", zap.String("subscriber_id", id))
	s.broadcaster.Register(id, ch)
	defer func() {
		s.logger.Info("subscriber disconnected", zap.String("subscriber_id", id))
		s.broadcaster.Unregister(id)
	}()

	for {
		select {
		case msg := <-ch:
			if err := stream.Send(msg); err != nil {
				s.logger.Error("failed to send metrics to subscriber",
					zap.String("subscriber_id", id),
					zap.Error(err),
				)
				return err
			}
			s.logger.Debug("sent metrics to subscriber", zap.String("subscriber_id", id))
		case <-stream.Context().Done():
			s.logger.Info("subscriber context canceled", zap.String("subscriber_id", id))
			return nil
		}
	}
}
