"use client";

import { createContext, useContext, useState, type ReactNode } from "react";

export type Locale = "en" | "zh";

const translations = {
  en: {
    // Header
    title: "IoT Configuration",
    subtitle: "SCADA Data Acquisition System",

    // Navigation
    models: "Models",
    settings: "Settings",
    tags: "Tags",
    commands: "Commands",
    events: "Events",

    // Model
    modelList: "Model List",
    addModel: "Add Model",
    editModel: "Edit Model",
    deleteModel: "Delete Model",
    modelId: "Model ID",
    driver: "Driver",
    groupId: "Group ID",
    noModels: "No models configured",

    // Device
    deviceList: "Device List",
    addDevice: "Add Device",
    editDevice: "Edit Device",
    deleteDevice: "Delete Device",
    deviceId: "Device ID",
    deviceName: "Device Name",
    deviceStatus: "Status",
    enabled: "Enabled",
    disabled: "Disabled",
    online: "Online",
    offline: "Offline",
    noDevices: "No devices in this model",

    // Settings
    connectionSettings: "Connection Settings",
    ipAddress: "IP Address",
    port: "Port",
    interval: "Polling Interval (s)",
    timeout: "Timeout (ms)",
    retries: "Retries",
    autoAddr: "Auto Address",
    unit: "Unit ID",

    // Tags
    tagList: "Tag List",
    addTag: "Add Tag",
    editTag: "Edit Tag",
    deleteTag: "Delete Tag",
    tagId: "Tag ID",
    tagName: "Tag Name",
    area: "Area",
    dataType: "Data Type",
    offset: "Offset",
    policy: "Policy",
    errView: "Error Display",
    noTags: "No tags configured",

    // Tag Areas
    coil: "Coil (0x)",
    discreteInput: "Discrete Input (1x)",
    inputRegister: "Input Register (3x)",
    holdingRegister: "Holding Register (4x)",

    // Data Types
    boolean: "Boolean",
    int16: "Int16",
    int16BE: "Int16 BE",
    int32: "Int32",
    int32BE: "Int32 BE",
    uint16: "UInt16",
    uint16BE: "UInt16 BE",
    uint32: "UInt32",
    uint32BE: "UInt32 BE",
    float32: "Float32",
    float32BE: "Float32 BE",
    float64: "Float64",
    float64BE: "Float64 BE",
    string: "String",

    // Commands
    commandList: "Command List",
    addCommand: "Add Command",
    editCommand: "Edit Command",
    deleteCommand: "Delete Command",
    commandName: "Command Name",
    defaultValue: "Default Value",
    ioWay: "I/O Mode",
    param: "Parameter",
    noCommands: "No commands configured",

    // Events
    eventList: "Event List",
    addEvent: "Add Event",
    editEvent: "Edit Event",
    deleteEvent: "Delete Event",
    eventName: "Event Name",
    condition: "Condition",
    action: "Action",
    noEvents: "No events configured",

    // Form
    form: "Form Schema",
    operations: "Operations",
    writeOut: "Write Output",

    // Actions
    save: "Save",
    cancel: "Cancel",
    delete: "Delete",
    confirm: "Confirm",
    export: "Export",
    import: "Import",
    duplicate: "Duplicate",

    // Messages
    confirmDelete: "Are you sure you want to delete this item?",
    saveSuccess: "Saved successfully",
    deleteSuccess: "Deleted successfully",
    error: "An error occurred",
    required: "This field is required",
    invalidIp: "Invalid IP address",
    invalidPort: "Port must be between 1 and 65535",
    duplicateId: "This ID already exists",

    // Search
    search: "Search...",
    filter: "Filter",

    // Policies
    policySave: "Save",
    policyNoSave: "Don't Save",
    policyOnChange: "On Change",

    // Common
    noSchema: "No configuration schema provided",
    devices: "devices",
    status: "Status",
  },
  zh: {
    // Header
    title: "物联网配置",
    subtitle: "SCADA 数据采集系统",

    // Navigation
    models: "模型",
    settings: "设置",
    tags: "点位",
    commands: "指令",
    events: "事件",

    // Model
    modelList: "模型列表",
    addModel: "添加模型",
    editModel: "编辑模型",
    deleteModel: "删除模型",
    modelId: "模型 ID",
    driver: "驱动",
    groupId: "分组 ID",
    noModels: "暂无配置模型",

    // Device
    deviceList: "设备列表",
    addDevice: "添加设备",
    editDevice: "编辑设备",
    deleteDevice: "删除设备",
    deviceId: "设备 ID",
    deviceName: "设备名称",
    deviceStatus: "状态",
    enabled: "启用",
    disabled: "禁用",
    online: "在线",
    offline: "离线",
    noDevices: "该模型下暂无设备",

    // Settings
    connectionSettings: "连接设置",
    ipAddress: "IP 地址",
    port: "端口",
    interval: "轮询间隔 (秒)",
    timeout: "超时 (毫秒)",
    retries: "重试次数",
    autoAddr: "自动寻址",
    unit: "单元 ID",

    // Tags
    tagList: "点位列表",
    addTag: "添加点位",
    editTag: "编辑点位",
    deleteTag: "删除点位",
    tagId: "点位 ID",
    tagName: "点位名称",
    area: "区域",
    dataType: "数据类型",
    offset: "偏移量",
    policy: "策略",
    errView: "错误显示",
    noTags: "暂无配置点位",

    // Tag Areas
    coil: "线圈 (0x)",
    discreteInput: "离散输入 (1x)",
    inputRegister: "输入寄存器 (3x)",
    holdingRegister: "保持寄存器 (4x)",

    // Data Types
    boolean: "布尔值",
    int16: "16位整数",
    int16BE: "16位整数 (大端)",
    int32: "32位整数",
    int32BE: "32位整数 (大端)",
    uint16: "16位无符号整数",
    uint16BE: "16位无符号整数 (大端)",
    uint32: "32位无符号整数",
    uint32BE: "32位无符号整数 (大端)",
    float32: "32位浮点数",
    float32BE: "32位浮点数 (大端)",
    float64: "64位浮点数",
    float64BE: "64位浮点数 (大端)",
    string: "字符串",

    // Commands
    commandList: "指令列表",
    addCommand: "添加指令",
    editCommand: "编辑指令",
    deleteCommand: "删除指令",
    commandName: "指令名称",
    defaultValue: "默认值",
    ioWay: "读写模式",
    param: "参数",
    noCommands: "暂无配置指令",

    // Events
    eventList: "事件列表",
    addEvent: "添加事件",
    editEvent: "编辑事件",
    deleteEvent: "删除事件",
    eventName: "事件名称",
    condition: "条件",
    action: "动作",
    noEvents: "暂无配置事件",

    // Form
    form: "表单结构",
    operations: "操作",
    writeOut: "写出配置",

    // Actions
    save: "保存",
    cancel: "取消",
    delete: "删除",
    confirm: "确认",
    export: "导出",
    import: "导入",
    duplicate: "复制",

    // Messages
    confirmDelete: "确定要删除此项吗？",
    saveSuccess: "保存成功",
    deleteSuccess: "删除成功",
    error: "发生错误",
    required: "此字段为必填项",
    invalidIp: "无效的 IP 地址",
    invalidPort: "端口必须在 1 到 65535 之间",
    duplicateId: "此 ID 已存在",

    // Search
    search: "搜索...",
    filter: "筛选",

    // Policies
    policySave: "保存",
    policyNoSave: "不保存",
    policyOnChange: "变化时保存",

    // Error Views
    showError: "显示错误",
    hideError: "不显示当前值",
    showLastValue: "显示上次值",

    // Common
    noSchema: "暂无配置 Schema",
    devices: "个设备",
    status: "状态",
  },
};

type TranslationKey = keyof (typeof translations)["en"];

interface I18nContextType {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: TranslationKey) => string;
}

const I18nContext = createContext<I18nContextType | undefined>(undefined);

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState<Locale>("zh");

  const t = (key: TranslationKey): string => {
    return translations[locale][key] || key;
  };

  return (
    <I18nContext.Provider value={{ locale, setLocale, t }}>
      {children}
    </I18nContext.Provider>
  );
}

export function useI18n() {
  const context = useContext(I18nContext);
  if (!context) {
    throw new Error("useI18n must be used within an I18nProvider");
  }
  return context;
}
