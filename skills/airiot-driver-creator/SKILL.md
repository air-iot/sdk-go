---
name: airiot-driver-creator
description: "AirIoT driver management tool with complete list of 82 drivers. Automatically match, install, and configure AirIoT drivers based on data collection requirements. When users describe requirements like collect modbus tcp data or connect opcua server, find the driver from references/drivers.md to get the correct Install Command, run it, then use the generated directory path to create data.json config file from schema.js. Critical mapping rules: driver.properties.settings maps to device.settings, model.properties.settings/tags/commands map to tables[].device settings/tags/commands, device.properties.settings maps to tables[].devices[].settings."
---

# AirIoT Driver Manager

AirIoT driver management tool for automatically matching, installing, and configuring AirIoT IoT device drivers based on user requirements.

## Directory Structure

Directory structure after driver installation:

```
airiot/lib/driver/
└── <actual-generated-dir>/    # Directory created after running npx install
    └── extracted/
        ├── schema.js      # Driver config structure (read-only)
        └── data.json      # Driver runtime config (to be generated)
```

**Key Paths:**
- `schema.js` location: `airiot/lib/driver/<actual-dir>/extracted/schema.js`
- `data.json` location: `airiot/lib/driver/<actual-dir>/extracted/data.json`

**NOTE:** The actual directory name is created automatically after running the npx install command. Use `ls` or check the actual directory to confirm the path.

## Usage Scenarios

When users describe data collection requirements, automatically match and install from the driver list:

**Example workflow:**
1. User: "Collect modbus tcp data at 127.0.0.1 port 502 unit 1 offset 200"
2. Match driver ID: `modbus` (from keyword matching)
3. Get install command from [references/drivers.md](references/drivers.md): `npx kesimodbus`
4. Run install command
5. Check actual generated directory (e.g., `airiot/lib/driver/modbus/extracted/`)
6. Read schema: `airiot/lib/driver/<actual-dir>/extracted/schema.js`
7. Generate config: `airiot/lib/driver/<actual-dir>/extracted/data.json`

## Workflow

1. **Parse requirement**: Extract device type/protocol from user description
2. **Match driver**: Find corresponding driver from [references/drivers.md](references/drivers.md)
3. **Get install command**: Get the **Install Command** from the table
4. **Install driver** (if needed): Run the install command: `npx kesi<package>`
5. **Check actual directory**: Use `ls airiot/lib/driver/` to find the actual generated directory name
6. **Read schema**: Read `schema.js` to understand config structure
7. **Generate config**: Create `data.json` based on user requirements

**IMPORTANT:** Always check the actual directory name after installation. The generated directory name may differ from the Driver ID used for matching.

### Generating data.json - Critical Mapping Rules

**⚠️ CRITICAL: Required Fields Checklist**

必填字段规则（简单记忆：**tables/devices/tags 需要 id+name，commands 需要 name**）：

| 层级 | 必填字段 | 示例 |
|------|----------|------|
| **tables[]** | `id`, `name` | `"modbus2"`, `"数据表1"` |
| **tables[].devices[]** | `id`, `name` | `"device-001"`, `"设备1"` |
| **tables[].device.tags[]** | `id`, `name` | `"tag-001"`, `"温度"` |
| **tables[].device.commands[]** | `name` | `"写入指令"` |
| **tables[].devices[].device.tags[]** | `id`, `name` | `"tag-001"`, `"温度"` |
| **tables[].devices[].device.commands[]** | `name` | `"写入指令"` |

**⚠️ CRITICAL: devices Array MUST Contain Devices**

`tables[].devices` 数组**不能为空**，必须至少包含一个设备：

```json
{
  "tables": [{
    "id": "table-001",
    "name": "数据表1",
    "device": { ... },
    "devices": [
      {
        "id": "device-001",       // 必填
        "name": "设备1",          // 必填
        "device": {
          "settings": { ... },
          "tags": [],
          "commands": []
        }
      }
    ]
  }]
}
```

**❌ 错误示例 - devices数组为空**:
```json
{
  "tables": [{
    "devices": []    // ❌ 错误：必须包含至少一个设备
  }]
}
```

**❌ 错误示例 - 缺少必填字段**:
```json
{
  "tables": [
    {
      // ❌ 缺少: "id" 和 "name"
      "device": { ... }
    }
  ]
}
```
    }
  ]
}
```

**Mapping Rules** (DO NOT copy schema.js structure directly):

| schema.js | data.json |
|-----------|-----------|
| `driver.properties.settings` | Root `device.settings` object |
| `model.properties.settings` | `tables[].device.settings` object |
| `model.properties.tags` | `tables[].device.tags` array |
| `model.properties.commands` | `tables[].device.commands` array |
| `device.properties.settings` | `tables[].devices[].device.settings` object |
| `device.properties.tags` | `tables[].devices[].device.tags` array |
| `device.properties.commands` | `tables[].devices[].device.commands` array |

**Example**:
```javascript
// schema.js structure
{
  driver: { properties: { settings: {...} } },
  model: { properties: { settings: {...}, tags: [...], commands: [...] } },
  device: { properties: { settings: {...} } }
}

// data.json structure (NOT the same as schema!)
{
  device: { settings: {...} },           // From driver.properties
  tables: [{
    id: "table-001",                     // REQUIRED
    name: "Table 1",                     // REQUIRED
    device: {                            // From model.properties
      settings: {...},
      tags: [...],
      commands: [...]
    },
    devices: [{                          // From device.properties
      id: "device-001",                  // REQUIRED
      name: "Device 1",                  // REQUIRED
      device: {    
         settings: {...},
         tags: [...],
         commands: [...]
       }
    }]
  }]
}
```

For detailed mapping examples, see [references/schema-guide.md](references/schema-guide.md) and [references/data-example.md](references/data-example.md).

**⚠️ CRITICAL: Validation Required**

After generating `data.json`, always validate:
1. All required fields are present (use the Required Fields Checklist)
2. Schema.js → data.json mapping rules are correctly followed
3. No direct copying of schema.js structure (driver/model/device top-level keys should not exist)

See [Data.json Validation](references/data-example.md) section for complete verification checklist.

## Common Driver Quick Match

**Keywords → Driver ID mapping for quick matching:**

| User Keywords | Driver ID |
|---------------|-----------|
| **PLC协议** | |
| modbus, mb, rtu, tcp | `modbus` |
| siemens, s7, 西门子 | `siemens-s7` |
| s7200, smart200, 西门子200 | `siemens-s7-200s` |
| mitsubishi, 三菱, mc | `mitsubishi-mc` |
| omron, 欧姆龙 | `omron` |
| ab, allen, bradley | `abplc` |
| **OPC协议** | |
| opcua, ua | `opcua` |
| opcda, da, opc | `opcda` |
| **物联网协议** | |
| mqtt, 消息队列 | `driver-mqtt-client` |
| lwm2m, 轻量级 | `driver-lwm2m-server` |
| coap, 物联网 | `driver-coap` |
| onenet, 中移, 移远 | `driver-onenet-mqtt` |
| **电力/水利协议** | |
| iec104, 104, 电力 | `iec104` |
| iec61850, 61850, 变电站 | `iec61850-mms-driver` |
| 651, 水利, 水文 | `sl651-2014-driver` |
| **楼宇/消防** | |
| bacnet, 楼宇, 暖通 | `bacnet` |
| tls350, 消防 | `tls350-driver` |
| **视频/摄像头** | |
| hik, 海康, 摄像头, 监控 | `driver-hik-camera` |
| ezviz, 萤石, 萤石云 | `driver-ezviz` |
| gb28181, 国标视频 | `driver-media-server` |
| camera, 摄像头, 通用 | `driver-common-camera` |
| **网络通信** | |
| tcp, tcp客户端 | `tcp-client-driver` |
| tcp, tcp服务端 | `tcp-server-driver` |
| udp, udp客户端 | `udp-client-driver` |
| udp, udp服务端 | `udp-server-driver` |
| websocket, ws, 客户端 | `websocket-client-driver` |
| websocket, ws, 服务端 | `websocket-server-driver` |
| http, http客户端 | `driver-http-client` |
| http, http服务端 | `http-server-driver` |
| kafka, 消息队列 | `kafka-driver` |
| snmp, 网管 | `driver-snmp` |
| **数据库** | |
| db, database, 数据库 | `db-driver` |
| mysql, oracle, postgres | `db-driver` |
| **其他设备** | |
| 报警, 周界 | `driver-ttzj` |
| 充电桩, 充电 | `driver-charging-kjtc` |
| 路灯, 照明 | `driver-lamp-sz` |
| led, 显示屏 | `driver-led-lednets` |

**Search Strategy:**
1. **Direct keyword match** - Look for exact protocol/device names
2. **Manufacturer match** - Siemens (西门子), Mitsubishi (三菱), Omron (欧姆龙), Hik (海康)
3. **Protocol type** - TCP/UDP/HTTP/MQTT/Modbus/OPC
4. **Industry domain** - 电力 (104/61850), 水利 (651), 楼宇 (BACnet), 消防 (TLS350)

See [references/drivers.md](references/drivers.md) for the complete list of 82 drivers with install commands.

## Resources

- **Complete driver list**: See [references/drivers.md](references/drivers.md) - Full mapping table of 82 drivers
- **Schema reference**: See [references/schema-guide.md](references/schema-guide.md) - Schema structure guide
- **Config examples**: See [references/data-example.md](references/data-example.md) - data.json examples