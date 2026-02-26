# Data.json Configuration Examples

`data.json` is the driver's runtime configuration file, generated based on `schema.js`.

## ⚠️ CRITICAL: Required Fields Checklist

**必填字段规则（简单记忆：tables/devices/tags 需要 id+name，commands 需要 name）**

| 层级 | 必填字段 | 说明 | 示例 |
|------|----------|------|------|
| **tables[]** | `id`, `name` | 每个table必须 | `"modbus2"`, `"数据表1"` |
| **tables[].devices[]** | `id`, `name` | 每个device必须 | `"device-001"`, `"设备1"` |
| **tables[].device.tags[]** | `id`, `name` | 每个tag必须 | `"tag-001"`, `"温度"` |
| **tables[].device.commands[]** | `name` | 每个command必须 | `"写入指令"` |
| **tables[].devices[].device.tags[]** | `id`, `name` | 每个tag必须 | `"tag-001"`, `"温度"` |
| **tables[].devices[].device.commands[]** | `name` | 每个command必须 | `"写入指令"` |

**⚠️ devices数组不能为空**

`tables[].devices` 数组**必须至少包含一个设备**：

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

**⚠️ COMMON MISTAKE - Copying schema structure**:
```json
// ❌ WRONG - This is schema.js structure, NOT data.json!
{
  "driver": { "settings": { ... } },
  "model": { "settings": { ... }, "tags": [ ... ], "commands": [] },
  "device": { "settings": { ... } }
}
```

## Schema to data.json Mapping

**CRITICAL MAPPING - Must Follow**:

| schema.js Section | data.json Location | Contains |
|-------------------|-------------------|----------|
| `driver.properties.settings` | Root `device.settings` | Driver-level default config (IP, port, interval, etc.) |
| `model.properties.settings` | `tables[].device.settings` | Model-level settings for each table |
| `model.properties.tags` | `tables[].device.tags` | Data point array for each table |
| `model.properties.commands` | `tables[].device.commands` | Command array for each table |
| `device.properties.settings` | `tables[].devices[].device.settings` | Device instance settings |
| `device.properties.tags` | `tables[].devices[].device.tags` | Device instance tags |
| `device.properties.commands` | `tables[].devices[].device.commands` | Device instance commands |

**Visual Representation**:
```
schema.js                      →  data.json
─────────────────────────────────────────────────────────────────────
driver.properties.settings     →  {
                                  "device": {
                                    "settings": { ... }  ← From driver.properties
                                  }
                                }

model.properties                →  {
                                  "tables": [{
                                    "device": {
                                      "settings": {},  ← From model.properties.settings
                                      "tags": [],     ← From model.properties.tags
                                      "commands": []  ← From model.properties.commands
                                    }
                                  }]
                                }

device.properties               →  {
                                  "tables": [{
                                    "devices": [        ← Array of device instances
                                      {
                                        "id": "...",
                                        "name": "...",
                                        "device": {
                                          "settings": {},  ← From device.properties.settings
                                          "tags": [],      ← From device.properties.tags
                                          "commands": []   ← From device.properties.commands
                                        }
                                      }
                                    ]
                                  }]
                                }
```

**Key Rule**: Do NOT copy schema structure directly. Map each schema section to its corresponding location in data.json.

## Structure Overview

```json
{
  "id": "设备ID",
  "name": "modbus",
  "groupId": "分组ID",
  "driverType": "modbus",
  "runMode": "one",
  "device": {
    "settings": {
      "interval": 60,
      "ip": "设备IP",
      "port": 502,
      "unit": 1
    }
  },
  "autoReload": {
    "disable": true
  },
  "autoUpdateConfig": true,
  "tables": [
    {
      "id": "modbus2",
      "name": "数据表1",
      "device": {
        "settings": {
          "autoAddr": true,
          "interval": 5,
          "ip": "设备IP"
        },
        "tags": [
          {
            "id": "p1",
            "name": "寄存器1",
            "area": 3,
            "dataType": "Int32BE",
            "offset": 1,
            "policy": "save"
          }
        ],
        "commands": [
          {
            "name": "写入指令",
            "ops": [
              {
                "area": "coil",
                "dataType": "Boolean",
                "offset": 1,
                "param": "cmd1"
              }
            ]
          }
        ]
      },
      "devices": [
        {
          "id": "device-001",
          "name": "设备1",
          "device": {
            "settings": {
              "unit": 7
            },
            "tags": [],
            "commands": []
          }
        }
      ]
    }
  ]
}
```

## Configuration Structure

| Layer | Location | Purpose | Required Fields |
|-------|----------|---------|-----------------|
| Top-level `device` | Root | Driver default config | - |
| `tables[]` | Root array | Table instances | `id`, `name`, `device` |
| `tables[].device` | Inside each table | Model config - defines data points template | - |
| `tables[].devices[]` | Inside each table | Device instances array | `id`, `name`, `device` |
| `tables[].devices[].device` | Inside each device instance | Device instance config with settings/tags/commands | - |

**Key Point**:
- `tables` array items **must** have `id` and `name` fields
- `devices` array items **must** have `id`, `name`, and `device` fields
- `tags` array items **must** have `id` and `name` fields
- `commands` array items **must** have `name` field

## Top-Level Fields

| Field | Description | Example |
|-------|-------------|---------|
| `id` | Driver instance ID | `"modbus-instance-001"` |
| `name` | Driver name | `"modbus"` |
| `groupId` | Group ID | `"group-001"` |
| `driverType` | Driver type | `"modbus"` |
| `runMode` | Run mode | `"one"` |
| `autoReload` | Auto reload config | `{ "disable": true }` |
| `autoUpdateConfig` | Auto update config | `true` |

## tables[] Configuration

Each table in the `tables` array **must** include:

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| `id` | Yes | Table identifier | `"modbus-table-001"` |
| `name` | Yes | Table name | `"数据表1"` |
| `device` | Yes | Model configuration object | See below |

### tables[].device Configuration

Model configuration with data points:

| Field | Description | Example |
|-------|-------------|---------|
| `settings.ip` | Device IP | `"192.168.1.100"` |
| `settings.port` | Device port | `502` |
| `settings.interval` | Collection period (seconds) | `5` |
| `settings.autoAddr` | Auto address | `true` |
| `tags` | **Data point array** | See below |
| `commands` | Command array | `[]` |

## Data Point Configuration (tags)

**⚠️ CRITICAL: Each tag MUST have both `id` and `name` fields!**

```json
{
  "id": "p1",              // ✅ REQUIRED - Tag identifier (MUST come first)
  "name": "寄存器1",        // ✅ REQUIRED - Tag name
  "area": 3,               // 1=coil, 2=input, 3=holding register, 4=input register
  "offset": 1,
  "dataType": "Int32BE",
  "policy": "save"
}
```

## tables[].devices[] Configuration

Device instances array inside each table. **Each device object must include**:

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| `id` | Yes | Device identifier | `"device-001"` |
| `name` | Yes | Device name | `"设备1"` |
| `device` | Yes | Device instance config object | See below |

### tables[].devices[].device Configuration

Each device instance has its own `device` object with:

| Field | Description | Example |
|-------|-------------|---------|
| `settings` | Device-specific settings (overrides model settings) | `{ "unit": 7 }` |
| `tags` | Device-specific tags (overrides model tags) | Array of tag objects |
| `commands` | Device-specific commands (overrides model commands) | Array of command objects |

```json
{
  "devices": [
    {
      "id": "device-001",      // ✅ REQUIRED
      "name": "设备1",          // ✅ REQUIRED
      "device": {
        "settings": {          // Device-specific settings
          "unit": 7
        },
        "tags": [],            // Device-specific tags
        "commands": []         // Device-specific commands
      }
    }
  ]
}
```

## Common Modbus Data Types

| dataType | Description |
|----------|-------------|
| `Boolean` | Boolean value |
| `Int16BE` | 16-bit signed integer (big-endian) |
| `Int16LE` | 16-bit signed integer (little-endian) |
| `UInt16BE` | 16-bit unsigned integer (big-endian) |
| `Int32BE` | 32-bit signed integer (big-endian) |
| `FloatBE` | 32-bit floating point (big-endian) |
| `String` | String |

## Data.json Validation

**验证生成的data.json是否正确遵循schema.js映射规则**

### 验证检查清单

在生成data.json后，使用此清单验证格式是否正确：

| # | 检查项 | 验证方法 | 示例 |
|---|--------|----------|------|
| 1 | **根结构正确性** | 不应包含`driver`/`model`/`device`顶级键 | ❌ `{"driver": {...}}` |
| 2 | **tables[]必需字段** | 每个table必须有`id`和`name` | ✅ `{"id": "t1", "name": "表1", ...}` |
| 3 | **devices[]必需字段** | 每个device必须有`id`、`name`、`device` | ✅ `{"id": "d1", "name": "设备1", "device": {...}}` |
| 4 | **devices[]非空检查** | `tables[].devices`数组必须至少包含一个设备 | ❌ `"devices": []` |
| 5 | **tags[]必需字段** | 每个tag必须有`id`和`name` | ✅ `{"id": "tag1", "name": "温度", ...}` |
| 6 | **commands[]必需字段** | 每个command必须有`name` | ✅ `{"name": "写入指令", ...}` |
| 7 | **device层级正确** | `tables[].devices[]`中每个元素必须有`device`包裹 | ✅ `{"device": {"settings": {...}}}` |
| 8 | **mapping规则遵守** | 检查schema.js各部分是否映射到正确位置 | 见下表 |

### Schema.js → data.json 转换验证示例

**假设schema.js结构**:
```javascript
// schema.js (只读参考文件)
{
  "driver": {
    "properties": {
      "settings": {
        "title": "驱动配置",
        "type": "object",
        "properties": {
          "ip": { "type": "string", "title": "IP地址" },
          "port": { "type": "number", "title": "端口" },
          "timeout": { "type": "number", "title": "超时时间" }
        }
      }
    }
  },
  "model": {
    "properties": {
      "settings": {
        "title": "模型配置",
        "properties": {
          "interval": { "type": "number", "title": "采集周期" }
        }
      },
      "tags": {
        "type": "array",
        "items": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "offset": { "type": "number" },
          "dataType": { "type": "string" }
        }
      },
      "commands": {
        "type": "array",
        "items": {
          "name": { "type": "string" }
        }
      }
    }
  },
  "device": {
    "properties": {
      "settings": {
        "title": "设备配置",
        "properties": {
          "unit": { "type": "number", "title": "从站地址" }
        }
      }
    }
  }
}
```

**正确的data.json输出** (基于上述schema.js):
```json
{
  "id": "driver-instance-001",
  "name": "驱动名称",
  "device": {
    "settings": {
      "ip": "192.168.1.100",
      "port": 502,
      "timeout": 5000
    },
    "tags": [],
    "commands": []
  },
  "tables": [
    {
      "id": "table-001",
      "name": "数据表1",
      "device": {
        "settings": {
          "interval": 5
        },
        "tags": [
          {
            "id": "tag-001",
            "name": "温度",
            "offset": 100,
            "dataType": "Int16BE"
          }
        ],
        "commands": [
          {
            "name": "写入指令"
          }
        ]
      },
      "devices": [
        {
          "id": "device-001",
          "name": "设备1",
          "device": {
            "settings": {
              "unit": 1
            },
            "tags": [],
            "commands": []
          }
        }
      ]
    }
  ]
}
```

### 映射验证表

| schema.js路径 | data.json路径 | 验证点 |
|---------------|---------------|--------|
| `driver.properties.settings` | `device.settings` | ✅ 根级device包含settings |
| `model.properties.settings` | `tables[].device.settings` | ✅ 每个table的device包含settings |
| `model.properties.tags` | `tables[].device.tags` | ✅ 每个table的device包含tags数组 |
| `model.properties.commands` | `tables[].device.commands` | ✅ 每个table的device包含commands数组 |
| `device.properties.settings` | `tables[].devices[].device.settings` | ✅ 每个devices元素的device包含settings |

### 常见错误验证

| 错误类型 | ❌ 错误示例 | ✅ 正确示例 |
|----------|------------|------------|
| **直接复制schema结构** | `{"driver": {...}, "model": {...}}` | `{"device": {...}, "tables": [...]}` |
| **缺少table必需字段** | `{"device": {...}}` | `{"id": "t1", "name": "表1", "device": {...}}` |
| **缺少device包裹层** | `{"devices": [{"settings": {...}}]}` | `{"devices": [{"device": {"settings": {...}}}]}` |
| **缺少tag必需字段** | `{"tags": [{"offset": 100}]}` | `{"tags": [{"id": "t1", "name": "温度", "offset": 100}]}` |
| **缺少command必需字段** | `{"commands": [{"ops": [...]}]}` | `{"commands": [{"name": "写入指令", "ops": [...]}]}` |

### 验证脚本逻辑

生成data.json时，确保执行以下验证：

```javascript
// 伪代码验证逻辑
function validateDataJson(data) {
  // 1. 检查顶级结构
  if (data.driver || data.model) {
    return "错误：不应包含driver/model顶级键，这是schema.js结构！";
  }

  // 2. 检查tables
  for (let table of data.tables) {
    if (!table.id || !table.name) {
      return `错误：table缺少id或name字段`;
    }

    // 3. 检查table的tags
    for (let tag of table.device.tags || []) {
      if (!tag.id || !tag.name) {
        return `错误：tag缺少id或name字段`;
      }
    }

    // 4. 检查table的commands
    for (let cmd of table.device.commands || []) {
      if (!cmd.name) {
        return `错误：command缺少name字段`;
      }
    }

    // 5. 检查devices非空
    if (!table.devices || table.devices.length === 0) {
      return `错误：devices数组不能为空，必须包含至少一个设备`;
    }

    // 6. 检查devices
    for (let device of table.devices) {
      if (!device.id || !device.name || !device.device) {
        return `错误：device缺少id、name或device对象`;
      }

      // 7. 检查device的tags
      for (let tag of device.device.tags || []) {
        if (!tag.id || !tag.name) {
          return `错误：device tag缺少id或name字段`;
        }
      }

      // 8. 检查device的commands
      for (let cmd of device.device.commands || []) {
        if (!cmd.name) {
          return `错误：device command缺少name字段`;
        }
      }
    }
  }

  return "✅ 验证通过";
}
```

## Complete Example

User requirement: "Collect modbus tcp at 127.0.0.1 port 502 unit 1 offset 200"

```json
{
  "id": "modbus-instance-001",
  "name": "modbus",
  "groupId": "group-001",
  "driverType": "modbus",
  "runMode": "one",
  "device": {
    "settings": {
      "interval": 60,
      "ip": "127.0.0.1",
      "port": 502,
      "unit": 1
    },
      "tags": [],
      "commands": []
  },
  "autoReload": {
    "disable": true
  },
  "autoUpdateConfig": true,
  "tables": [
    {
      "id": "modbus2",
      "name": "数据表1",
      "device": {
        "settings": {
          "autoAddr": true,
          "interval": 5,
          "ip": "127.0.0.1",
          "port": 502,
          "unit": 1
        },
        "tags": [
          {
            "id": "reg200",
            "name": "寄存器200",
            "area": 3,
            "offset": 200,
            "dataType": "Int16BE",
            "policy": "save"
          }
        ],
        "commands": []
      },
      "devices": [
        {
          "id": "device-001",
          "name": "设备1",
          "device": {
            "settings": {
              "unit": 1
            },
            "tags": [],
            "commands": []
          }
        }
      ]
    }
  ]
}
```