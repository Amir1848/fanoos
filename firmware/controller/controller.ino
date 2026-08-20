/*
 * Fanoos ESP32 controller firmware.
 *
 * One board plays two roles over a single MQTT connection:
 *   - the controller itself (reports momentary button "press" events)
 *   - every light circuit wired to it (drives its light, accepts cmd, reports state)
 *
 * Speak exactly the contract in docs/architucture.md §8:
 *   state: home/{env}/device/{guid}/state  (retained, QoS 1)
 *   cmd:    home/{env}/device/{guid}/cmd
 * Last Will on the controller's state topic reports the whole board offline.
 *
 * Library: PubSubClient by Nick O'Leary (Arduino Library Manager).
 */

#include <WiFi.h>
#include <PubSubClient.h>
#include <time.h>

// ============================================================================
//  USER CONFIGURATION — edit these to match your network
// ============================================================================
static const char* WIFI_SSID     = "honor";
static const char* WIFI_PASSWORD = "amir1843";

// Address of the Mosquitto broker the backend talks to.
// For a local docker-compose broker on the same machine, use the host PC's
// LAN IP (not "localhost" — that's the ESP32 itself).
static const char* MQTT_BROKER_HOST = "192.168.43.173";
static const uint16_t MQTT_BROKER_PORT = 1883;
// ============================================================================

// ---------------------------------------------------------------------------
//  Hardcoded device identity (looked up from the database once, baked in).
//  env=4  controller=e1095eea-0c1f-4b4f-8ed5-82b9cd464764  ("کنترلر 1")
//  light =2728e8e8-9ab4-4bcc-9480-8b273b87e613            ("چراغ 1")
//
//  "چراغ 1" is mapped to the ESP32's built-in LED, so the board can be turned
//  on/off from the server with no external hardware wired up.
// ---------------------------------------------------------------------------
static const char* ENVIRONMENT_ID   = "4";
static const char* CONTROLLER_GUID  = "e1095eea-0c1f-4b4f-8ed5-82b9cd464764";

// On most ESP32 dev boards the built-in LED is on GPIO 2 and is active-HIGH.
// (Hardcoded because not every ESP32 board variant defines LED_BUILTIN.)
static const uint8_t BUILTIN_LED_PIN = 2;

struct Circuit {
  const char* guid;
  const char* name;
  uint8_t     lightPin;     // output that drives the light
  bool        activeHigh;   // true: HIGH = on (LED); false: LOW = on (relay)
  uint8_t     buttonPin;    // optional local toggle button (INPUT_PULLUP)
  bool        on;
  bool        lastButtonState;
  bool        lastPublishedOn;
};

// To add another light later, copy this block and assign free GPIOs.
// Use activeHigh=true for LEDs, activeHigh=false for active-LOW relay modules.
static Circuit CIRCUITS[] = {
  {
    "2728e8e8-9ab4-4bcc-9480-8b273b87e613",
    "چراغ 1",
    BUILTIN_LED_PIN,   // built-in LED output (GPIO 2)
    true,              // built-in LED is active-HIGH
    5,                 // optional toggle button in (leave unwired if unused)
    false, false, false,
  },
};
static const uint8_t CIRCUIT_COUNT = sizeof(CIRCUITS) / sizeof(CIRCUITS[0]);

// Controller's own momentary button.
static const uint8_t CONTROLLER_BUTTON_PIN = 4;

static const char* WIFI_HOSTNAME = "fanoos-controller";
static const char* MQTT_CLIENT_ID = "esp32-e1095eea";

static const uint32_t DEBOUNCE_MS        = 50;
static const uint32_t CONTROLLER_PULSE_MS = 150;   // matches cmd/simulator
static const uint32_t MQTT_RECONNECT_MS  = 2000;

static WiFiClient    net;
static PubSubClient  mqtt(net);

// Cached time so NTP failure doesn't break reporting.
static uint32_t nowUnix() {
  return (uint32_t)time(nullptr);
}

static uint32_t bootEpochOverride() {
  // Fallback when NTP hasn't synced yet: pretend boot was unix 0 + elapsed.
  return millis() / 1000;
}

static uint32_t tsNow() {
  uint32_t t = nowUnix();
  return t > 1700000000 ? t : bootEpochOverride();
}

// ---------------------------------------------------------------------------
//  JSON payloads are tiny; snprintf keeps the firmware dependency-free.
// ---------------------------------------------------------------------------
static void sendControllerState(bool pressed) {
  char topic[80];
  char payload[160];
  snprintf(topic, sizeof(topic), "home/%s/device/%s/state", ENVIRONMENT_ID, CONTROLLER_GUID);
  snprintf(payload, sizeof(payload),
           "{\"type\":\"controller\",\"pressed\":%s,\"online\":true,\"ts\":%lu}",
           pressed ? "true" : "false",
           (unsigned long)tsNow());
  mqtt.publish(topic, payload, true);   // retained
}

static void sendCircuitState(Circuit& c) {
  char topic[80];
  char payload[160];
  snprintf(topic, sizeof(topic), "home/%s/device/%s/state", ENVIRONMENT_ID, c.guid);
  snprintf(payload, sizeof(payload),
           "{\"type\":\"light\",\"on\":%s,\"online\":true,\"ts\":%lu}",
           c.on ? "true" : "false",
           (unsigned long)tsNow());
  mqtt.publish(topic, payload, true);   // retained
  c.lastPublishedOn = c.on;
}

static void applyLight(Circuit& c) {
  // active-HIGH:  on -> HIGH   (LED)
  // active-LOW :  on -> LOW    (most relay modules)
  digitalWrite(c.lightPin, (c.on == c.activeHigh) ? HIGH : LOW);
}

// ---------------------------------------------------------------------------
//  MQTT
// ---------------------------------------------------------------------------
static void onMqttMessage(char* topic, byte* payload, unsigned int len) {
  // Look for the light's GUID inside the cmd topic it subscribed to.
  for (uint8_t i = 0; i < CIRCUIT_COUNT; i++) {
    Circuit& c = CIRCUITS[i];

    char cmdTopic[80];
    snprintf(cmdTopic, sizeof(cmdTopic),
             "home/%s/device/%s/cmd", ENVIRONMENT_ID, c.guid);
    if (strcmp(topic, cmdTopic) != 0) continue;

    // Body looks like:  {"on":true}   or   {"on":false}
    // Find "on" then read the next token after ':'.
    char buf[64];
    if (len >= sizeof(buf)) len = sizeof(buf) - 1;
    memcpy(buf, payload, len);
    buf[len] = '\0';

    char* p = strstr(buf, "\"on\"");
    if (!p) return;
    p = strchr(p + 4, ':');
    if (!p) return;
    p++;
    while (*p == ' ' || *p == '\t') p++;

    bool target = (*p == 't');   // "true"
    c.on = target;
    applyLight(c);
    Serial.printf("[cmd] %s -> on=%s\n", c.name, c.on ? "true" : "false");
    sendCircuitState(c);
    return;
  }
}

static void subscribeCircuits() {
  for (uint8_t i = 0; i < CIRCUIT_COUNT; i++) {
    Circuit& c = CIRCUITS[i];
    char cmdTopic[80];
    snprintf(cmdTopic, sizeof(cmdTopic),
             "home/%s/device/%s/cmd", ENVIRONMENT_ID, c.guid);
    mqtt.subscribe(cmdTopic);
  }
}

static void announceAllOnline() {
  sendControllerState(false);
  for (uint8_t i = 0; i < CIRCUIT_COUNT; i++) {
    sendCircuitState(CIRCUITS[i]);
  }
}

static bool connectMqtt() {
  char willTopic[80];
  char willPayload[64];
  snprintf(willTopic, sizeof(willTopic),
           "home/%s/device/%s/state", ENVIRONMENT_ID, CONTROLLER_GUID);
  snprintf(willPayload, sizeof(willPayload),
           "{\"online\":false,\"ts\":%lu}", (unsigned long)tsNow());

  Serial.printf("[mqtt] connecting to %s:%u as %s ...\n",
                MQTT_BROKER_HOST, MQTT_BROKER_PORT, MQTT_CLIENT_ID);
  bool ok = mqtt.connect(MQTT_CLIENT_ID,
                         nullptr, nullptr,       // no user/pass (broker allows anonymous)
                         willTopic, 1, true, willPayload);
  if (ok) {
    Serial.println("[mqtt] connected");
  } else {
    Serial.printf("[mqtt] connect failed, state=%d\n", mqtt.state());
  }
  return ok;
}

// ---------------------------------------------------------------------------
//  Buttons
// ---------------------------------------------------------------------------
static bool edgeLow(bool& last, bool reading) {
  bool fell = (last && !reading);
  last = reading;
  return fell;
}

// Controller button = momentary: edge -> publish pressed:true, pulse, then false.
static bool controllerBtnLast = true;       // pull-up: idle = HIGH
static bool controllerPulsing = false;
static uint32_t controllerPulseStart = 0;
static bool controllerPublishedTrue = false;

static void readControllerButton() {
  bool reading = digitalRead(CONTROLLER_BUTTON_PIN) == LOW;
  bool fell    = edgeLow(controllerBtnLast, reading);
  if (!fell) return;

  sendControllerState(true);
  controllerPublishedTrue = true;
  controllerPulsing = true;
  controllerPulseStart = millis();
}

static void tickControllerPulse() {
  if (!controllerPulsing) return;
  if (millis() - controllerPulseStart < CONTROLLER_PULSE_MS) return;
  sendControllerState(false);
  controllerPulsing = false;
  controllerPublishedTrue = false;
}

// Light toggle button = push-on / push-off: edge toggles the light.
struct ButtonDebounce {
  bool last;
  uint32_t lastChange;
};
static ButtonDebounce circuitBtn[sizeof(CIRCUITS) / sizeof(CIRCUITS[0])] = {};

static void readCircuitButtons() {
  uint32_t now = millis();
  for (uint8_t i = 0; i < CIRCUIT_COUNT; i++) {
    Circuit& c = CIRCUITS[i];
    ButtonDebounce& db = circuitBtn[i];
    bool reading = digitalRead(c.buttonPin) == LOW;
    if (reading == db.last) continue;
    if (now - db.lastChange < DEBOUNCE_MS) continue;
    db.last = reading;
    db.lastChange = now;
    if (!reading) continue;            // only toggle on press (HIGH->LOW edge)

    c.on = !c.on;
    applyLight(c);                     // local action always wins — works offline
    if (mqtt.connected()) {
      sendCircuitState(c);
    }
  }
}

// ---------------------------------------------------------------------------
//  WiFi
// ---------------------------------------------------------------------------
static uint32_t lastWifiAttempt = 0;

static void connectWifi() {
  if (WiFi.status() == WL_CONNECTED) return;
  uint32_t now = millis();
  // Don't call WiFi.begin() every loop — restarting mid-handshake causes the
  // "connects then disconnects" cycle. Throttle like the MQTT reconnect.
  if (now - lastWifiAttempt < 5000) return;
  lastWifiAttempt = now;
  Serial.printf("[wifi] connecting to \"%s\" ...\n", WIFI_SSID);
  WiFi.hostname(WIFI_HOSTNAME);
  WiFi.disconnect();
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
}

// ---------------------------------------------------------------------------
//  Arduino entry points
// ---------------------------------------------------------------------------
void setup() {
  Serial.begin(115200);
  delay(100);
  Serial.println("\n[boot] fanoos controller starting");
  Serial.printf("[boot] env=%s controller=%s circuits=%u\n",
                ENVIRONMENT_ID, CONTROLLER_GUID, CIRCUIT_COUNT);

  pinMode(CONTROLLER_BUTTON_PIN, INPUT_PULLUP);
  for (uint8_t i = 0; i < CIRCUIT_COUNT; i++) {
    Circuit& c = CIRCUITS[i];
    pinMode(c.lightPin, OUTPUT);
    pinMode(c.buttonPin, INPUT_PULLUP);
    applyLight(c);
    circuitBtn[i].last = digitalRead(c.buttonPin) == LOW;
    circuitBtn[i].lastChange = 0;
  }

  // NTP, soft: doesn't block if broker has no internet.
  configTime(0, 0, "pool.ntp.org");

  WiFi.mode(WIFI_STA);
  connectWifi();

  mqtt.setServer(MQTT_BROKER_HOST, MQTT_BROKER_PORT);
  mqtt.setKeepAlive(15);
  mqtt.setBufferSize(256);
  mqtt.setCallback(onMqttMessage);
}

static uint32_t lastMqttAttempt = 0;

void loop() {
  // Drive WiFi + MQTT in the foreground; no blocking delays in the hot path.
  static bool wifiWasUp = false;
  if (WiFi.status() != WL_CONNECTED) {
    if (wifiWasUp) {
      Serial.println("[wifi] disconnected");
      wifiWasUp = false;
    }
    connectWifi();
  } else {
    if (!wifiWasUp) {
      Serial.printf("[wifi] connected, IP=%s\n", WiFi.localIP().toString().c_str());
      wifiWasUp = true;
    }
    if (!mqtt.connected()) {
      if (millis() - lastMqttAttempt >= MQTT_RECONNECT_MS) {
        lastMqttAttempt = millis();
        if (connectMqtt()) {
          subscribeCircuits();
          announceAllOnline();
        }
      }
    } else {
      mqtt.loop();
    }
  }

  readControllerButton();
  readCircuitButtons();
  tickControllerPulse();

  // Publish light state if it changed while MQTT was briefly down.
  if (mqtt.connected()) {
    for (uint8_t i = 0; i < CIRCUIT_COUNT; i++) {
      Circuit& c = CIRCUITS[i];
      if (c.lastPublishedOn != c.on) {
        sendCircuitState(c);
      }
    }
  }
}
