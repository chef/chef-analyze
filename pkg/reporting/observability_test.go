package reporting

import (
	"testing"
)

type recordingSink struct {
	logs    []string
	metrics []string
}

func (r *recordingSink) Log(event string, keyValues ...string) {
	r.logs = append(r.logs, event)
}

func (r *recordingSink) Metric(name string, value int64, keyValues ...string) {
	r.metrics = append(r.metrics, name)
}

func TestSetObservabilitySinkAndEmit(t *testing.T) {
	sink := &recordingSink{}
	SetObservabilitySink(sink)
	t.Cleanup(func() {
		SetObservabilitySink(nil)
	})

	emitLog("reporting.test.log", "k", "v")
	emitMetric("reporting.test.metric", 1, "k", "v")

	if len(sink.logs) != 1 || sink.logs[0] != "reporting.test.log" {
		t.Fatalf("unexpected logs: %#v", sink.logs)
	}
	if len(sink.metrics) != 1 || sink.metrics[0] != "reporting.test.metric" {
		t.Fatalf("unexpected metrics: %#v", sink.metrics)
	}
}

func TestSetObservabilitySinkNilIsNoop(t *testing.T) {
	SetObservabilitySink(nil)
	emitLog("reporting.test.noop")
	emitMetric("reporting.test.noop.metric", 1)
}
