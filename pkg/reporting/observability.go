package reporting

import "sync"

// ObservabilitySink receives structured reporting logs and metrics.
// It is optional; the default sink is a no-op.
type ObservabilitySink interface {
	Log(event string, keyValues ...string)
	Metric(name string, value int64, keyValues ...string)
}

type noopObservabilitySink struct{}

func (noopObservabilitySink) Log(string, ...string)           {}
func (noopObservabilitySink) Metric(string, int64, ...string) {}

var (
	obsMu   sync.RWMutex
	obsSink ObservabilitySink = noopObservabilitySink{}
)

// SetObservabilitySink configures package-level logging/metrics emission.
// Passing nil resets behavior to no-op.
func SetObservabilitySink(sink ObservabilitySink) {
	obsMu.Lock()
	defer obsMu.Unlock()
	if sink == nil {
		obsSink = noopObservabilitySink{}
		return
	}
	obsSink = sink
}

func emitLog(event string, keyValues ...string) {
	obsMu.RLock()
	sink := obsSink
	obsMu.RUnlock()
	sink.Log(event, keyValues...)
}

func emitMetric(name string, value int64, keyValues ...string) {
	obsMu.RLock()
	sink := obsSink
	obsMu.RUnlock()
	sink.Metric(name, value, keyValues...)
}
