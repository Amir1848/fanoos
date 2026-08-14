# ESP32 Controller Firmware

Real firmware for the controller board, replacing `cmd/simulator` for in hardware
testing. Implements the same MQTT contract documented in
[`docs/architucture.md`](../../docs/architucture.md) §8, so the backend talks to
real hardware exactly as it talks to the simulator.

## Hardware

- ESP32 dev board (any variant with at least 4 free GPIOs).
- One momentary pushbutton for the controller (`CONTROLLER_BUTTON_PIN`).
- Per light circuit:
  - One push-to-toggle button (`Circuit.buttonPin`) — optional, leave unwired if unused.
  - One light output (`Circuit.lightPin`).

By default, **"چراغ 1" is mapped to the board's built-in LED** (`LED_BUILTIN`,
GPIO 2 on most ESP32 dev boards) so it can be turned on/off from the server with
no external hardware wired up. The optional local toggle button needs nothing
wired (its `INPUT_PULLUP` reads idle HIGH), so server-only control works on a
bare board.

Default wiring (edit the constants in `controller.ino` to match your board):

| Function              | GPIO          | Notes                                 |
| --------------------- | ------------- | ------------------------------------- |
| Controller button     | 4             | Momentary, INPUT\_PULLUP, active LOW  |
| Light 1 toggle button | 5             | INPUT\_PULLUP, active LOW (optional)  |
| Light 1 output        | `LED_BUILTIN` | Built-in LED, active-HIGH             |

Buttons idle HIGH (internal pull-up); a press pulls them LOW.

Each circuit carries its own polarity in `activeHigh`: `true` for LEDs
(HIGH = on), `false` for active-LOW relay modules (LOW = on). To drive a relay
instead of the built-in LED, set `lightPin` to the relay GPIO and
`activeHigh = false`.

## Configuration

Open `controller.ino` and edit the top block:

```cpp
static const char* WIFI_SSID     = "YOUR_WIFI_SSID";
static const char* WIFI_PASSWORD = "YOUR_WIFI_PASSWORD";
static const char* MQTT_BROKER_HOST = "192.168.1.10";  // your PC's LAN IP
static const uint16_t MQTT_BROKER_PORT = 1883;
```

Use the **LAN IP of the machine running Mosquitto** (`docker compose up
mosquitto`), not `localhost` — that would resolve to the ESP32 itself.

The device identity block below (`ENVIRONMENT_ID`, `CONTROLLER_GUID`, the
`CIRCUITS[]` table) is already hardcoded to your database (env 4, controller
`e1095eea-...`, light `2728e8e8-...`).

## Flash

1. Install [Arduino IDE](https://www.arduino.cc/en/software) or use
   `arduino-cli`.
2. Add ESP32 board support (Boards Manager → search "esp32").
3. Install **PubSubClient** by Nick O'Leary via Library Manager.
4. Open `controller.ino`, select your ESP32 board + port, upload.

## Behavior

- On boot: connects WiFi, then MQTT with a Last Will of
  `{"online":false}` on the controller's state topic. Once connected,
  publishes `online:true` for the controller and every light (retained),
  and subscribes to each light's `cmd` topic.
- Controller button press → emits the momentary `pressed:true` →
  `pressed:false` transition (150 ms pulse, same as `cmd/simulator`).
- Light toggle button → flips the light locally and publishes the new
  state. Local action always wins and works even with the network down
  (per project principle: manual control has priority).
- Light `cmd` `{"on":bool}` from the backend → drives the light and
  publishes the resulting state.

## Multiple lights

The `CIRCUITS[]` array makes adding lights a copy-paste job — add another
`{ guid, name, lightPin, activeHigh, buttonPin, false, false, false }` row. The MQTT
contract already gives each light its own `cmd`/`state` topics under its own
GUID, so no backend change is needed.

## Time source

`ts` in payloads uses NTP (`pool.ntp.org`) with a fallback to
seconds-since-boot if NTP hasn't synced. The backend treats `ts` as a freshness
hint, not a validated value, so either is acceptable.
