# Security Policy & Architecture - PrefixOS

This document specifies the security architecture, threat model, authentication/authorization semantics, TLS/mTLS setup, and vulnerability disclosure procedures for PrefixOS.

---

## 1. Threat Model & Security Boundaries

PrefixOS operates as an in-cluster control plane component. Security boundaries protect against unauthorized token inspection, cache poisoning, denial-of-service (DoS) memory exhaustion, and unauthenticated node join attacks.

```
Client Traffic ──> [ TLS 1.3 / mTLS / API Key Auth ] ──> [ PrefixOS Control Plane Engine ]
                                                                   │
                                                        [ mTLS Node Replication ]
                                                                   v
                                                        [ Cluster Raft Peers ]
```

---

## 2. Security Controls

- **TLS 1.3 / mTLS**: All gRPC and REST HTTP communications mandate TLS 1.3 encryption. Node-to-node Raft replication uses mutual TLS (mTLS) with client certificate verification.
- **API Key & JWT Authentication**: Requests MUST carry a valid `Authorization: Bearer <token>` header or `X-API-Key` matching configured tenant keys.
- **Rate Limiting**: Token-bucket rate limiter enforces request limits per tenant (`requests_per_sec`).
- **Vulnerability Disclosure**: To report security issues, email `security@prefixos.io`.
