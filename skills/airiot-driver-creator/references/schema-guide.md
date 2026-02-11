# Schema Structure Guide

The `schema.js` file of an AirIoT driver describes the configuration structure of the driver.

## Schema Structure

```javascript
{
  "driver": {
    "properties": {
      "settings": {
        "title": "Driver Config",
        "type": "object",
        "properties": { ... }
      }
    }
  },
  "model": {
    "properties": {
      "settings": { ... },
      "tags": { ... },
      "commands": { ... }
    }
  },
  "device": {
    "properties": {
      "settings": { ... }
    }
  }
}
```

## Schema to data.json Mapping

**Critical Mapping Rules**:

| schema.js Section | data.json Location | Description |
|-------------------|-------------------|-------------|
| `driver.properties.settings` | Root `device.settings` | Driver-level default configuration (IP, port, interval, etc.) |
| `model.properties.*` | `tables[].device` | Model configuration for each table (settings, tags, commands) |
| `device.properties.*` | `tables[].devices[]` | Device instance configuration (each device is an array item) |

**Visual Mapping**:
```
schema.js                          →  data.json
─────────────────────────────────────────────────────────────────────────────
driver.properties.settings         →  { "device": { "settings": {...} } }
model.properties.settings          →  { "tables": [{ "device": { "settings": {...} } }] }
model.properties.tags              →  { "tables": [{ "device": { "tags": [...] } }] }
model.properties.commands          →  { "tables": [{ "device": { "commands": [...] } }] }
device.properties                  →  { "tables": [{ "devices": [{ ... }] }] }
```

## Main Configuration Sections

### driver.properties - Driver Configuration

**Maps to**: `data.json` root `device.settings`

Driver-level common configuration, typically includes:
- `ip`: Device IP address
- `port`: Port number
- `timeout`: Connection timeout
- `interval`: Collection period

### model.properties - Model Configuration

**Maps to**: `data.json` `tables[]`

Each table object contains:
- `id`: Table identifier (required)
- `name`: Table name (required)
- `device`: Model configuration object

#### Table device object structure

- `model.properties.settings` → `device.settings` - Model-level default configuration
- `model.properties.tags` → `device.tags` - Data point array
- `model.properties.commands` → `device.commands` - Write command array

#### Data Point Configuration (tags)

Each data point contains:
- `name`: Name
- `id`: Identifier
- `area`: Read area (1=coil status, 2=input status, 3=holding register, 4=input register)
- `offset`: Offset address
- `dataType`: Data type (Int16BE, FloatBE, Boolean, etc.)
- `policy`: Storage policy (e.g., "save")

#### Command Configuration (commands)

Array of write commands for sending control commands to the device.

### device.properties - Device Configuration

**Maps to**: `data.json` `tables[].devices[]`

Device-specific configuration that can override model settings. Each object in `devices` array represents one device instance and **must contain**:
- `id`: Device identifier (required)
- `name`: Device name (required)
- `settings`: Device-specific settings (optional)