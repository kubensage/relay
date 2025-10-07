package grpc

import (
	"sync"

	"github.com/kubensage/relay/proto/gen"
	"go.uber.org/zap"
)

// Broadcaster manages a set of subscribers and allows broadcasting
// metrics to all active listeners concurrently.
//
// Each subscriber is identified by an ID and associated with a channel.
// Broadcasts are non-blocking: if a subscriber's channel is full, the
// message is dropped to avoid stalling other subscribers.
type Broadcaster struct {
	subscribersMu sync.RWMutex                 // Protects concurrent access to subscribers
	subscribers   map[string]chan *gen.Metrics // Map of subscriber ID to metrics channel
	logger        *zap.Logger                  // Logger for observability
}

// NewBroadcaster creates and returns a new Broadcaster.
//
// Parameters:
//   - logger: zap.Logger for observability (can be nil).
//
// Returns:
//   - *Broadcaster: a new Broadcaster instance.
func NewBroadcaster(logger *zap.Logger) *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[string]chan *gen.Metrics),
		logger:      logger,
	}
}

// Register adds a new subscriber with the given ID and metrics channel.
//
// Parameters:
//   - id: Unique subscriber identifier.
//   - ch: Channel where metrics will be delivered.
func (b *Broadcaster) Register(id string, ch chan *gen.Metrics) {
	b.subscribersMu.Lock()
	defer b.subscribersMu.Unlock()
	b.subscribers[id] = ch

	if b.logger != nil {
		b.logger.Info("subscriber registered", zap.String("id", id))
	}
}

// Unregister removes the subscriber associated with the given ID.
//
// Parameters:
//   - id: Identifier of the subscriber to remove.
func (b *Broadcaster) Unregister(id string) {
	b.subscribersMu.Lock()
	defer b.subscribersMu.Unlock()
	delete(b.subscribers, id)

	if b.logger != nil {
		b.logger.Info("subscriber unregistered", zap.String("id", id))
	}
}

// Broadcast delivers a metrics message to all active subscribers.
//
// Behavior:
//   - If the subscriber's channel has capacity, the message is sent.
//   - If the channel is full, the message is dropped and a warning is logged.
//
// Parameters:
//   - msg: Metrics message to broadcast.
func (b *Broadcaster) Broadcast(msg *gen.Metrics) {
	b.subscribersMu.RLock()
	defer b.subscribersMu.RUnlock()

	for id, ch := range b.subscribers {
		select {
		case ch <- msg:
			if b.logger != nil {
				b.logger.Debug("broadcasted message", zap.String("subscriber_id", id))
			}
		default:
			if b.logger != nil {
				b.logger.Warn("dropping metrics: subscriber channel full", zap.String("subscriber_id", id))
			}
		}
	}
}
