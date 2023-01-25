package helpers

import (
	"bytes"
	"encoding/binary"
	"strconv"

	pdk "github.com/extism/go-pdk"
)

/*
#include "../../runtime/extism-pdk.h"
*/
import "C"

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

// Make these an import
func SetError(err error) {
	if err == nil {
		return
	}

	SetErrorString(err.Error())
	return
}

func SetErrorString(err string) {
	mem := pdk.AllocateString(err)
	defer mem.Free()
	C.extism_error_set(mem.Offset())
}

func IncrNumberKey(key string) error {
	if key == "" {
		return nil
	}

	val := pdk.GetVar(key)
	if val == nil {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, 1)
		pdk.SetVar(key, b)

		return nil
	}

	intVal := binary.LittleEndian.Uint64(val)
	intVal = intVal + 1
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, intVal)
	pdk.SetVar(key, b)

	return nil
}

func GetKeyUint64(key string) uint64 {
	if key == "" {
		return 0
	}

	val := pdk.GetVar(key)
	if val == nil {
		return 0
	}

	intVal := binary.LittleEndian.Uint64(val)
	return intVal
}

func LogString(l pdk.LogLevel, s string) {
	mem := pdk.AllocateString(s)
	defer mem.Free()
	pdk.LogMemory(l, mem)
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
	incoming := pdk.GetVar(pingKey)
	if incoming == nil {
		LogString(pdk.LogDebug, "process_ guest: could not get pings recv")
		b.Write([]byte(`0,"pongs":`))
	} else {
		v := binary.LittleEndian.Uint64(incoming)
		b.Write([]byte(strconv.FormatUint(v, 10)))
		b.Write([]byte(`,"pongs":`))
	}
	outgoing := pdk.GetVar(pongKey)
	if outgoing == nil {
		LogString(pdk.LogDebug, "process_ guest: could not get pongs sent")
		b.Write([]byte(`0`))
	} else {
		v := binary.LittleEndian.Uint64(outgoing)
		b.Write([]byte(strconv.FormatUint(v, 10)))
	}

	b.Write([]byte(`,"plugin":"`))
	b.Write([]byte(plugin))
	b.Write([]byte(`","async":`))
	b.Write([]byte(strconv.FormatBool(async)))
	b.Write([]byte(`}`))

	return b
}
