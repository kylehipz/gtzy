# Blood-sugar meter sync over Bluetooth

gtzy can pull stored readings directly off an Accu-Chek Guide (or any glucose
meter that implements the standard Bluetooth Glucose Profile) and store them
alongside manually entered readings.

## How it works (and why it's a "pull", not a "push")

Bluetooth Low Energy is asymmetric. The **meter is the peripheral / GATT
server**; it holds the records. The phone app — and now gtzy — is the **central
/ client** that connects to the meter and *pulls* the stored records. gtzy is
**not** a server the meter pushes to; there is nothing to "receive" a
connection. `gtzy serve` must run on a machine that has a Bluetooth adapter and
is within range of the meter.

The meter speaks the **standard Bluetooth Glucose Profile** (it is
Continua-certified), so no proprietary decoding is involved:

| GATT | UUID | Role in sync |
|------|------|--------------|
| Glucose Service | `0x1808` | Container service |
| Glucose Measurement | `0x2A18` | Each stored reading, streamed as an indication |
| Record Access Control Point (RACP) | `0x2A52` | We write "report stored records" here |

Flow (`internal/meter/meter.go`): connect → discover the service → subscribe to
Glucose Measurement + RACP indications → write the RACP "report stored records"
command → the meter streams one indication per record → we decode each per the
spec (`internal/meter/decode.go`) → insert into `blood_sugar_readings`.

## Prerequisites

- Linux with **BlueZ** running (this is the backend `tinygo.org/x/bluetooth`
  uses; it talks to BlueZ over D-Bus, no CGO).
- A Bluetooth adapter on the machine running `gtzy serve`.
- `bluetoothctl` available (ships with BlueZ).

## One-time setup: pair the meter

gtzy only ever **connects** to an already-bonded meter — it does not run a BLE
pairing agent. You pair once, out of band, with `bluetoothctl`, which handles
the passkey prompt:

```sh
bluetoothctl
[bluetooth]# power on
[bluetooth]# agent on
[bluetooth]# scan on
# Put the meter into pairing/Bluetooth mode (see its manual). Watch for a line
# like:  [NEW] Device AA:BB:CC:DD:EE:FF  Accu-Chek ...
[bluetooth]# pair AA:BB:CC:DD:EE:FF
# Enter the passkey shown on the meter's screen when prompted.
[bluetooth]# trust AA:BB:CC:DD:EE:FF
[bluetooth]# scan off
[bluetooth]# exit
```

> Note the meter's MAC (`AA:BB:CC:DD:EE:FF`). Pairing with gtzy's machine is a
> *second, independent* bond that lives alongside the one your phone app uses;
> it does not remove the phone pairing.

### Tell gtzy which device to use

Set the meter's MAC so gtzy connects straight to it:

```sh
export GTZY_METER_ADDR=AA:BB:CC:DD:EE:FF
```

If `GTZY_METER_ADDR` is unset, gtzy falls back to scanning and connecting to the
first device whose advertised name contains "accu-chek". Setting the MAC is
faster and unambiguous, so prefer it.

## Running a sync

Start the server on the Bluetooth machine, then trigger a sync either way:

```sh
gtzy serve            # in one shell, on the machine with the adapter
gtzy sync             # in another shell — pulls new records
```

or click **Sync meter** on the Blood Sugar tab in the web UI. Both hit
`POST /api/bloodsugar/sync`.

Make sure the meter is awake and in range when you sync. The first sync imports
all stored records; later syncs only ask for records newer than the highest
sequence number already imported.

### Idempotent by design

Re-running sync never creates duplicates. Each meter record carries a sequence
number, stored in `blood_sugar_readings.seq_number`, and a partial unique index
(`idx_bsr_meter_seq`) with `ON CONFLICT DO NOTHING` drops any record already
present. Sync as often as you like.

## How meter records map to the database

Each Glucose Measurement record becomes one `blood_sugar_readings` row:

| DB column | Source |
|-----------|--------|
| `value_mgdl` | Decoded SFLOAT glucose concentration, converted to mg/dL |
| `taken_at` | Record base time + optional time offset (RFC3339) |
| `seq_number` | The meter's record sequence number (drives dedup) |
| `source` | `"meter"` |
| `meal_tag` | Empty — meal context is a separate BLE characteristic that gtzy does not read; tag readings manually in the UI if you want |
| `notes` | Empty |

### Verify units on your first sync (important, one time)

The decoder assumes the meter reports concentration in **kg/L** (Accu-Chek's
convention) and converts to mg/dL; the units flag also handles mol/L. Meters are
standard-compliant, but the safe check is: after the first sync, compare one
reading in gtzy against the same reading on the meter's screen.

- **They match** → the mapping is correct for your device; trust every future sync.
- **Off by a constant factor** → adjust the conversion constants in
  `internal/meter/decode.go` (`mgdlPerKgL` / `mgdlPerMolL`). The decode logic is
  isolated and unit-tested (`internal/meter/decode_test.go`), so this is a
  contained, low-risk change.

## Troubleshooting

- **`enable bluetooth adapter` error** — BlueZ isn't running or the adapter is
  off. `bluetoothctl power on`, and check `systemctl status bluetooth`.
- **Connect times out / "is it bonded and awake?"** — the meter is asleep, out
  of range, or not bonded. Wake it, move closer, and confirm it appears under
  `bluetoothctl` → `devices`.
- **Won't connect while the phone app is active** — some meters allow only one
  active link at a time. Briefly close/disconnect the phone app during the sync;
  bonds are independent, so re-pairing is not needed.
- **Sync returns "no glucose meter found"** — `GTZY_METER_ADDR` is unset and the
  name scan didn't match. Set `GTZY_METER_ADDR` to the meter's MAC.
- **Values look wrong** — see "Verify units" above.
