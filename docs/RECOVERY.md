# Disaster Recovery Runbook - PrefixOS

This runbook documents emergency recovery procedures for node crashes, corrupted WAL logs, network partition healing, and snapshot restores.

---

## 1. Cold Start Snapshot Recovery Procedure

1. Stop the PrefixOS service: `systemctl stop prefixos`
2. Verify latest valid snapshot: `ls -lh /var/lib/prefixos/snapshots/`
3. Execute parallel startup recovery command:
   `./prefixos --config=configs/config.yaml --recover-snapshot=snap-latest.bin`
4. Inspect startup logs to confirm recovery time (`< 5.0 seconds`).

---

## 2. WAL Corruption Recovery

If CRC32 checksum fails during startup:
1. PrefixOS automatically halts startup to prevent state pollution.
2. Truncate unchecksummed log tail:
   `./tools/wal-repair --wal-file=/var/lib/prefixos/wal/prefixos.wal`
3. Restart server: `systemctl start prefixos`

---

## 3. WAL Log Rotation & Checkpoint Truncation Policy

To maintain fast recovery times as workload volume grows:
1. **Periodic Checkpoint**: The `SnapshotManager` periodically creates a CoW binary snapshot.
2. **WAL Truncation**: Upon successful snapshot creation, WAL entries prior to the snapshot's max sequence ID are safely truncated.
3. **Replay Efficiency**: Node restart replays only records newer than the latest valid snapshot ID.

