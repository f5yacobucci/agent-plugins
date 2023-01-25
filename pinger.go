package main

import (
	"fmt"
	"strconv"

	"github.com/f5yacobucci/agent-plugins/internal/helpers"

	pdk "github.com/extism/go-pdk"
	"github.com/valyala/fastjson"
)

/*
#include "runtime/extism-pdk.h"
*/
import "C"

//go:wasm-module env
//export process__
func process__(uint64, uint64, uint64, uint64) uint64

// with multi returns tinygo will encode return values into a pointer that's set during
// a host side call, therefore, the expected multi return:
//func process__(uint64, uint64, uint64, uint64) (int32, uint64, uint64)
// will compile to
// func process__(uint32, uint64, uint64, uint64, uint64)
// where the first parameter is a pointer to linear memory
// https://github.com/tinygo-org/tinygo/issues/3254
// single return value (returnable directly):
// func(uint64, uint64, uint64, uint64) int32 -> just return a single integer

// Plugin logic below
var (
	name    = ""
	version = ""
)

//export init_
func init_() int32 {
	// DEBUG Metrics
	err := helpers.IncrNumberKey(helpers.InvocationsInitTimes)
	if err != nil {
		helpers.SetError(err)
		return -1
	}

	// XXX figure out logging, try 0.2.0 release or git HEAD
	// XXX add the plugin name to each log
	helpers.LogString(pdk.LogDebug, "init_ guest: entry")

	helpers.LogString(pdk.LogDebug, "init_ guest: exit")
	return 0
}

//export close_
func close_() int32 {
	// DEBUG Metrics
	err := helpers.IncrNumberKey(helpers.InvocationsCloseTimes)
	if err != nil {
		helpers.SetError(err)
		return -1
	}

	helpers.LogString(pdk.LogDebug, "close_ guest: entry")

	helpers.LogString(pdk.LogDebug, "close_ guest: exit")
	return 0
}

//export process_
func process_() int32 {
	// DEBUG Metrics
	err := helpers.IncrNumberKey(helpers.InvocationsProcessTimes)
	if err != nil {
		helpers.SetError(err)
		return -1
	}

	helpers.LogString(pdk.LogDebug, "process_ guest: entry")

	name, ok := pdk.GetConfig(helpers.PluginName)
	if !ok {
		helpers.LogString(pdk.LogDebug, "process_ guest: cannot get self name")
		name = "unknown"
	}

	input := pdk.Input()

	var p fastjson.Parser
	v, err := p.ParseBytes(input)
	if err != nil {
		helpers.SetError(err)
		return -1
	}

	topic := v.GetStringBytes("topic")
	if topic == nil {
		helpers.SetErrorString("process_ guest: cannot determine the event topic")
		return -1
	}

	if string(topic) == helpers.Pong {
		helpers.IncrNumberKey(helpers.PongsRecv)
		if err != nil {
			helpers.LogString(pdk.LogDebug, "process_ guest: failed incrementing pongs")
		}

		limit, ok := pdk.GetConfig(helpers.Limit)
		if !ok {
			helpers.LogString(pdk.LogDebug, "process_ guest: cannot determine limit, stopping")
			return 0
		}

		n, err := strconv.ParseUint(limit, 10, 64)
		if err != nil {
			// call on_error?
			helpers.LogString(pdk.LogDebug, "process_ guest: cannot parse limit config, stopping")
			return 0
		}

		if helpers.GetKeyUint64(helpers.PongsRecv) == n {
			helpers.LogString(pdk.LogDebug, "process_ geust: limit reached")
			return 0
		}
	}

	ping := pdk.AllocateString(helpers.Ping)
	defer ping.Free()
	payload := pdk.AllocateString(name)
	defer payload.Free()

	ret := process__(
		ping.Offset(),
		ping.Length(),
		payload.Offset(),
		payload.Length(),
	)
	if ret > 0 {
		mem := pdk.FindMemory(ret)
		buf := make([]byte, mem.Length())
		mem.Load(buf)
		helpers.SetError(fmt.Errorf(
			"process_ guest: host side process__ failed - rc: %d, msg: %s",
			uint64(ret),
			string(buf),
		))
		return -1
	}

	err = helpers.IncrNumberKey(helpers.PingsSent)
	if err != nil {
		helpers.LogString(pdk.LogDebug, "process_ guest: failed incrementing pings")
	}
	helpers.LogString(pdk.LogDebug, "process_ guest: host side process__ success")

	b := helpers.BuildReturn(name, topic, false, helpers.PingsSent, helpers.PongsRecv)
	mem := pdk.AllocateBytes(b.Bytes())
	defer mem.Free()
	pdk.OutputMemory(mem)

	helpers.LogString(pdk.LogDebug, "process_ guest: exit")
	return 0
}

//export info_
func info_() int32 {
	// DEBUG Metrics
	err := helpers.IncrNumberKey(helpers.InvocationsInfoTimes)
	if err != nil {
		helpers.SetError(err)
		return -1
	}

	helpers.LogString(pdk.LogDebug, "info_ guest: entry")

	arena := &fastjson.Arena{}
	json := arena.NewObject()
	json.Set("name", arena.NewString(name))
	json.Set("version", arena.NewString(version))

	var b []byte
	enc := json.MarshalTo(b)

	mem := pdk.AllocateBytes(enc)
	defer mem.Free()
	pdk.OutputMemory(mem)

	helpers.LogString(pdk.LogDebug, "info_ guest: exit")
	return 0
}

//export subscriptions_
func subscriptions_() int32 {
	// DEBUG Metrics
	err := helpers.IncrNumberKey(helpers.InvocationsSubsTimes)
	if err != nil {
		helpers.SetError(err)
		return -1
	}

	helpers.LogString(pdk.LogDebug, "subscriptions_ guest: entry")

	subs := pdk.AllocateString(helpers.Subscriptions)
	defer subs.Free()
	pdk.OutputMemory(subs)

	helpers.LogString(pdk.LogDebug, "subscriptions_ guest: exit")
	return 0
}

// https://github.com/tinygo-org/tinygo/issues/2703
func main() {}
