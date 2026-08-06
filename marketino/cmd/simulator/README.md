# Device Simulator

Stands in for real ESP32 firmware, which doesn't exist yet. Speaks the MQTT
topic/payload contract documented in [`docs/architucture.md`](../../../docs/architucture.md) §8, so
backend MQTT integration can be developed and tested without hardware.

One process simulates one physical controller board: it reports its own
`press` events, and drives every light circuit wired to it — applying `cmd`
messages from the backend and reporting each circuit's resulting state.
Unlike real firmware, which never talks to the database directly (see
docs/architucture.md §7), the simulator reads its circuits from Postgres at
startup as a dev convenience.

## Run a local broker

```sh
docker compose up mosquitto
```

## Wire a light to a controller as a circuit

A light only gets simulated if there's a `device_links` row with
`link_type="circuit"`, `source_device_id` = the controller, and
`target_device_id` = the light:

```sh
curl -X POST localhost:3000/device-links/create \
  -H 'Content-Type: application/json' \
  -d '{"sourceDeviceId": <controller id>, "targetDeviceId": <light id>, "linkType": "circuit"}'
```

(The `/device-links/create` endpoint already accepts any `linkType` string —
no backend change needed. There's currently no UI for creating `circuit`
links, only the `toggle` links used for switch-controls-light behavior; add
one if this becomes a real feature rather than dev tooling.)

## Run a simulated controller

```sh
go run ./cmd/simulator -controller 4b162738-96b2-4492-85b2-dea7530d90b9
```

- `-controller` must match a real `devices.guid` row where `device_type` is
  Controller (2). The environment ID and every linked circuit are looked up
  from the database automatically — no need to pass them separately.
- Each linked light circuit gets its own `cmd`/`state` topics, exactly as if
  it were an independently connected light device (its own MQTT connection
  and Last Will, so it reports offline if the process dies uncleanly).
- `-broker` defaults to `tcp://localhost:1883`.

At the prompt, type:

- `press` + Enter — the controller's own momentary `pressed:true` →
  `pressed:false` transition.
- `<circuit-name> on` / `<circuit-name> off` + Enter — simulate a manual
  toggle at that circuit (`<circuit-name>` is the light's `devices.name`).

There is no backend-side mediation of `device_links` yet (e.g. turning a
`toggle`-linked controller's `press` into a `cmd` on its linked light) —
until that exists, drive it manually with `mosquitto_pub`/`mosquitto_sub`.
