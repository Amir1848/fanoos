# Device Simulator

Stands in for real ESP32 firmware, which doesn't exist yet. Speaks the MQTT
topic/payload contract documented in [`docs/architucture.md`](../../../docs/architucture.md) §8, so
backend MQTT integration can be developed and tested without hardware.

## Run a local broker

```sh
docker compose up mosquitto
```

## Run a simulated device

```sh
go run ./cmd/simulator -env 1 -device 1a3f3b0e-7d9c-48a9-98a2-cc756a1eebbf -type light
go run ./cmd/simulator -env 1 -device 4b162738-96b2-4492-85b2-dea7530d90b9 -type controller
```

- `-env` should match a real `environments.id` row.
- `-device` should match a real `devices.guid` row (returned from
  `POST /devices/create`, or looked up via the `all-devices` dataview) so
  topics line up with what the backend will publish/subscribe to. The
  integer `devices.id` is internal-only and not used for MQTT addressing.
- `-type light` subscribes to its `cmd` topic and republishes its `state` on
  every command (and once on startup).
- `-type controller` has no commands; type `press` + Enter at the prompt to
  publish a momentary `pressed:true` → `pressed:false` state transition.
- `-broker` defaults to `tcp://localhost:1883`.

There is no backend-side mediation of `device_links` yet (e.g. turning a
Controller's `press` into a `cmd` on its linked Light) — until that exists,
drive it manually with `mosquitto_pub`/`mosquitto_sub`, or run a second
simulator instance and bridge the two by hand.
