package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/f5yacobucci/agent-plugins/internal/helpers"

	"github.com/valyala/fastjson"
	wapc "github.com/wapc/wapc-guest-tinygo"
)

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

func init_(_ []byte) ([]byte, error) {
	// DEBUG Metrics
	helpers.IncrNumberKey(helpers.InvocationsInitTimes)

	_, _ = os.LookupEnv("HOME")

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

	if string(topic) == helpers.Pong {
		wapc.ConsoleLog("process_ guest: received pong event")
		helpers.IncrNumberKey(helpers.PongsRecv)

		/*
			limit, ok := pdk.GetConfig(helpers.Limit)
			if !ok {
				helpers.LogString(pdk.LogDebug, "process_ guest: cannot determine limit, stopping")
				return 0
			}
		*/
		var limit uint64 = 10 // this should be set from config

		if helpers.GetNumberKey(helpers.PongsRecv) == limit {
			wapc.ConsoleLog("process_ guest: limit reached")
			b := helpers.BuildReturn(name, topic, false, helpers.PingsSent, helpers.PongsRecv)
			return b.Bytes(), nil
		}
	}

	var msg bytes.Buffer
	msg.Write([]byte(`{"topic":"`))
	msg.Write([]byte(helpers.Ping))
	msg.Write([]byte(`","data":""}`))
	wapc.HostCall("pinger", "messagebus", "process__", msg.Bytes())

	helpers.IncrNumberKey(helpers.PingsSent)
	wapc.ConsoleLog("process_ guest: host side process__ success")

	b := helpers.BuildReturn(name, topic, false, helpers.PingsSent, helpers.PongsRecv)
	wapc.ConsoleLog(fmt.Sprintf("process_ guest output: %v", b.String()))

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
	// DEBUG Metrics
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
