all: ping pong

ping:
	tinygo build -opt=0 -ldflags="-X main.name=pinger -X main.version=v0.1.0" -o pinger.wasm -scheduler=none -target=wasi pinger.go

pong:
	tinygo build -opt=0 -ldflags="-X main.name=ponger -X main.version=v0.1.0" -o ponger.wasm -scheduler=none -target=wasi ponger.go
