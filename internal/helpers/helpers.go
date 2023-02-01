package helpers

import (
	"bytes"
	"strconv"
)

const (
	// topics - should pull these from agent, but CGO whines
	AgentStarted = "agent.started"
	Ping         = "nginx.plugin.external.ping"
	Pong         = "nginx.plugin.external.pong"

	Subscriptions = `["` + AgentStarted + `"]`

	// Keys
	InvocationsInitTimes    = "invocations.init.times"
	InvocationsSubsTimes    = "invocations.subs.times"
	InvocationsCloseTimes   = "invocations.close.times"
	InvocationsInfoTimes    = "invocations.info.times"
	InvocationsProcessTimes = "invocations.process.times"
	PingsSent               = "nginx.heartbeat.ping.sent"
	PingsRecv               = "nginx.heartbeat.ping.recv"
	PongsSent               = "nginx.heartbeat.pong.sent"
	PongsRecv               = "nginx.heartbeat.pong.recv"

	// Config Keys
	PluginName = "plugin-name"
	Limit      = "limit"
)

type metrics map[string]uint64

var (
	pluginMetrics metrics
)

func init() {
	pluginMetrics = make(metrics)
}

func IncrNumberKey(key string) {
	var v uint64
	var ok bool
	if v, ok = pluginMetrics[key]; !ok {
		pluginMetrics[key] = 1
		return
	}
	v = v + 1
	pluginMetrics[key] = v
}

func GetNumberKey(key string) uint64 {
	var v uint64
	var ok bool
	if v, ok = pluginMetrics[key]; !ok {
		return 0
	}
	return v
}

func BuildReturn(
	plugin string,
	topic []byte,
	async bool,
	pingKey string,
	pongKey string,
) bytes.Buffer {
	var b bytes.Buffer
	b.Write([]byte(`{"topic":"`))
	b.Write(topic)
	b.Write([]byte(`","pings":`))
	incoming := GetNumberKey(pingKey)
	b.Write([]byte(strconv.FormatInt(int64(incoming), 10)))
	b.Write([]byte(`,"pongs":`))
	outgoing := GetNumberKey(pongKey)
	b.Write([]byte(strconv.FormatInt(int64(outgoing), 10)))

	b.Write([]byte(`,"plugin":"`))
	b.Write([]byte(plugin))
	b.Write([]byte(`","async":`))
	b.Write([]byte(strconv.FormatBool(async)))
	b.Write([]byte(`}`))

	return b
}
