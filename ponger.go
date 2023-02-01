package main

import (
	"bytes"
	"fmt"

	"github.com/f5yacobucci/agent-plugins/internal/helpers"

	"github.com/valyala/fastjson"
	wapc "github.com/wapc/wapc-guest-tinygo"
)

// See pinger.go for notes on multivalue returns

// Plugin logic below
var (
	name    = ""
	version = ""
)

func init_(_ []byte) ([]byte, error) {
	helpers.IncrNumberKey(helpers.InvocationsInitTimes)

	wapc.ConsoleLog("init_ guest: entry")
	wapc.ConsoleLog(fmt.Sprintf("init_ guest invoked: %v", helpers.GetNumberKey(helpers.InvocationsInitTimes)))

	wapc.ConsoleLog("init_ guest: exit")
	return nil, nil
}

func close_(_ []byte) ([]byte, error) {
	// DEBUG Metrics
	helpers.IncrNumberKey(helpers.InvocationsCloseTimes)

	wapc.ConsoleLog("close_ guest: entry")
	wapc.ConsoleLog(fmt.Sprintf("close_ guest invoked: %v", helpers.GetNumberKey(helpers.InvocationsCloseTimes)))

	wapc.ConsoleLog("close_ guest: exit")
	return nil, nil
}

func process_(input []byte) ([]byte, error) {
	// DEBUG Metrics
	helpers.IncrNumberKey(helpers.InvocationsProcessTimes)

	wapc.ConsoleLog("process_ guest: entry")
	wapc.ConsoleLog(fmt.Sprintf("process_ guest invoked: %v", helpers.GetNumberKey(helpers.InvocationsProcessTimes)))

	/*
	   name, ok := pdk.GetConfig(helpers.PluginName)
	   if !ok {
	     helpers.LogString(pdk.LogDebug, "process_ guest: cannot get self name")
	     name = "unknown"
	   }
	*/

	var p fastjson.Parser
	v, err := p.ParseBytes(input)
	if err != nil {
		return nil, err
	}

	topic := v.GetStringBytes("topic")
	if topic == nil {
		return nil, err
	}

	if string(topic) == helpers.Ping {
		helpers.IncrNumberKey(helpers.PingsRecv)

		var msg bytes.Buffer
		msg.Write([]byte(`{"topic":"`))
		msg.Write([]byte(helpers.Pong))
		msg.Write([]byte(`","data":""}`))
		wapc.HostCall("ponger", "messagebus", "process__", msg.Bytes())

		helpers.IncrNumberKey(helpers.PongsSent)
		wapc.ConsoleLog("process_ guest: host side process__ success")
	}

	b := helpers.BuildReturn(name, topic, false, helpers.PingsRecv, helpers.PongsSent)

	wapc.ConsoleLog("process_ guest: exit")
	return b.Bytes(), nil
}

func info_(_ []byte) ([]byte, error) {
	// DEBUG Metrics
	helpers.IncrNumberKey(helpers.InvocationsInfoTimes)

	wapc.ConsoleLog("info_ guest: entry")
	wapc.ConsoleLog(fmt.Sprintf("info_ guest invoked: %v", helpers.GetNumberKey(helpers.InvocationsInfoTimes)))

	arena := &fastjson.Arena{}
	json := arena.NewObject()
	json.Set("name", arena.NewString(name))
	json.Set("version", arena.NewString(version))

	var b []byte
	enc := json.MarshalTo(b)

	wapc.ConsoleLog("info_ guest: exit")
	return enc, nil
}

func subscriptions_(_ []byte) ([]byte, error) {
	helpers.IncrNumberKey(helpers.InvocationsSubsTimes)

	wapc.ConsoleLog("subscriptions_ guest: entry")
	wapc.ConsoleLog(fmt.Sprintf("subscriptions_ guest invoked: %v", helpers.GetNumberKey(helpers.InvocationsSubsTimes)))

	wapc.ConsoleLog("subscriptions_ guest: exit")
	return []byte(helpers.Subscriptions), nil
}

// https://github.com/tinygo-org/tinygo/issues/2703
func main() {
	wapc.RegisterFunctions(wapc.Functions{
		"init_":          init_,
		"close_":         close_,
		"subscriptions_": subscriptions_,
		"info_":          info_,
		"process_":       process_,
	})
}
