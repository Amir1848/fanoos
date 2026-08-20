# Smart Home Automation Platform

## Overview

This project is an open-source home automation platform designed around a simple principle:

> **Make existing homes smart without changing the electrical wiring.**

Instead of replacing traditional switches with expensive smart switches, this project adds a low-cost controller behind the existing electrical panel (or junction box). The controller communicates with a central server and can control lights and other electrical devices while preserving normal manual operation.

The system is designed to be inexpensive, scalable, and easy to install.

---

## Goals

* Keep existing electrical wiring unchanged.
* Keep manual switches fully functional.
* Allow remote control from a server.
* Support automation scenarios.
* Minimize hardware cost.
* Be open-source and easily extensible.

---

## System Architecture

The system consists of two major components:

### 1. Central Server

The central server is responsible for:

* Device management
* User authentication
* Automation rules
* MQTT communication
* State synchronization
* APIs for mobile/web applications
* Natural-language assistant — LLM chat, device tool-calling, and push-to-talk voice input implemented (see architecture doc §9)

The server is implemented in **Go (Golang)**. See [`architucture.md`](architucture.md) for the concrete component/data-flow/database design.

### 2. Device Controller

Each home contains one or more low-cost controllers (ESP32 or similar).

Responsibilities include:

* Reading physical switch states
* Controlling relays
* Communicating with the server using MQTT
* Continuing local operation even if the server becomes unavailable

The controller should always prioritize safety and manual operation.

---

## Communication

Protocol:

* MQTT

Communication model:

```
ESP32 <------ MQTT ------> Go Server
```

The server acts as the system brain while controllers execute commands locally.

---

## Electrical Design Philosophy

One of the most important requirements of this project is compatibility with traditional electrical installations.

The system should work with existing switches, including:

* Standard switches
* Two-way (staircase) switches
* Future switch configurations

Manual operation must always remain available.

Even if:

* the Wi-Fi connection is lost,
* the MQTT broker becomes unavailable,
* the server is offline,

users must still be able to operate lights normally using physical switches.

---

## Design Principles

The project follows these principles:

1. Manual control is always the highest priority.
2. Smart features should never reduce system reliability.
3. Hardware should be inexpensive.
4. Installation should require minimal electrical modifications.
5. Components should be modular and replaceable.
6. All communication should be asynchronous.
7. Every controller should operate independently whenever possible.

---

## AI Context

This section is intended for AI assistants working on this repository.

When generating code or suggesting architectural improvements, always assume the following constraints:

* Existing home wiring must remain unchanged.
* Physical switches always have priority over remote commands.
* The ESP32 hardware has limited RAM and CPU resources.
* MQTT is the primary communication protocol.
* The Go server is responsible for orchestration, authentication, and automation logic.
* Network failures are expected and must be handled gracefully.
* Devices should continue operating locally during network outages.
* Safety and reliability are more important than adding new features.
* Backward compatibility is preferred over introducing breaking changes.

Whenever proposing new features, ensure they respect these constraints.

---

## Future Features

Planned capabilities include:

* Scheduling
* Automation rules
* **LLM-powered assistant** — users type or speak (push-to-talk, server-side whisper.cpp STT) commands like "turn on the kitchen light"; the server sends the text to a cheap tool-calling LLM (e.g. DeepSeek, or a local model for privacy), and the LLM invokes existing device APIs as tools (implemented — see architecture doc §9).
* Scene management
* Energy monitoring
* OTA firmware updates
* Device grouping
* Presence detection
* Voice assistant integration
* Local-first automation
* Multi-home support

---

## Project Status

🚧 Early Development

The architecture is still evolving. Documentation will continue to expand as hardware and software components mature.
