# Compatibility Matrix - PrefixOS

This matrix documents supported operating systems, Go versions, container runtimes, Kubernetes versions, and LLM inference engine frameworks for PrefixOS.

---

## 1. Runtime & Operating System Compatibility

| Subsystem | Requirement / Compatibility | Status |
| :--- | :--- | :--- |
| **Go Compiler** | `Go 1.22+` | ✅ Tested & Supported |
| **Linux OS** | Ubuntu 20.04+, RHEL 8+, Debian 11+ (x86_64 / arm64) | ✅ Primary Production Platform |
| **macOS** | macOS 12+ (Apple Silicon M1/M2/M3 / Intel) | ✅ Supported for local dev |
| **Windows OS** | Windows 10 / 11 / Server 2022 (x64) | ✅ Supported |
| **Docker** | Docker Engine `24.0+`, Docker Compose `v2.20+` | ✅ Verified |
| **Kubernetes** | Kubernetes `1.28+` / `1.30+` | ✅ Tested with StatefulSets |

---

## 2. LLM Inference Engine & Framework Support

| Framework | Integration Mode | Supported Versions | Status |
| :--- | :--- | :--- | :--- |
| **vLLM** | Direct gRPC / IPC Adapter | `v0.4.0+` / `v0.5.0+` | ✅ Native Adapter (`v2.5`) |
| **SGLang** | RadixAttention gRPC API | `v0.1.10+` | ✅ Native Adapter (`v2.5`) |
| **llama.cpp** | C++ CFFI / REST Proxy | `b2500+` | ✅ Native Adapter (`v2.5`) |
| **TensorRT-LLM** | gRPC Control Plane API | `v0.8.0+` | ✅ gRPC Stubs Supported |
| **OpenAI API Proxy** | REST Gateway | ChatCompletions API v1 | ✅ REST Adapter (`v1.0`) |
