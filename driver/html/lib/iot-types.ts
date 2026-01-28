// IoT Configuration Types based on Modbus TCP format

export interface Tag extends Record<string, unknown> {
  id: string;
  name: string;
}

export interface FormField {
  name: string;
  type: string;
  defaultValue?: Record<string, unknown>;
  ioway?: string;
  mod?: string | null;
  tag?: string | null;
  tagValue?: unknown;
  select?: unknown;
  select2?: unknown;
  arrayValue?: unknown;
  objectValue?: unknown;
  objectValue2?: unknown;
  tableValue?: unknown;
  tableValue2?: unknown;
  ifRepeat?: unknown;
}

export interface Operation {
  area: string;
  dataType: string;
  offset: number;
  param: string;
}

export interface WriteOut {
  mod?: string | null;
  tag?: string | null;
  tagValue?: unknown;
  select?: unknown;
  select2?: unknown;
  arrayValue?: unknown;
  objectValue?: unknown;
  objectValue2?: unknown;
  tableValue?: unknown;
  tableValue2?: unknown;
  ifRepeat?: unknown;
}

export interface Command extends Record<string, unknown> {
  name: string;
}

export interface Event extends Record<string, unknown> {
  name: string;
}

export interface DeviceSettings {
  driver: string;
  groupId: string;
  settings: Record<string, unknown>;
  tags: Tag[];
  commands: Command[];
  events: Event[];
}

export interface Device {
  id: string;
  name: string;
  device: DeviceSettings;
  _settings?: {
    device: DeviceSettings;
  };
  disable: boolean;
  off: boolean;
}

export interface ModelSettings {
  ip?: string;
  port?: number;
  interval?: number;
  timeout?: number;
  retries?: number;
  autoAddr?: boolean;
  [key: string]: unknown;
}

export interface ModelDevice {
  driver: string;
  groupId: string;
  settings: ModelSettings;
  tags: Tag[];
  commands: Command[];
  events: Event[];
}

export interface Model {
  id: string;
  name: string;
  device: DeviceSettings;
  devices: Device[];
}

// JSON Schema types for form generation
export interface JSONSchemaProperty {
  type?: "string" | "number" | "integer" | "boolean" | "array" | "object";
  title?: string;
  description?: string;
  default?: unknown;
  enum?: (string | number | boolean)[];
  enumNames?: string[];
  enum_title?: string[];
  minimum?: number;
  maximum?: number;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  format?: string;
  fieldType?: "password" | "textarea" | "deviceScriptEdit" | string;
  defaultScript?: string;
  items?: JSONSchemaProperty;
  properties?: Record<string, JSONSchemaProperty>;
  required?: string[];
}

export interface JSONSchema {
  type?: "object";
  title?: string;
  description?: string;
  properties: Record<string, JSONSchemaProperty>;
  required?: string[];
}

// Area options for Modbus
export const MODBUS_AREAS = [
  { value: 0, label: "coil" },
  { value: 1, label: "discreteInput" },
  { value: 3, label: "inputRegister" },
  { value: 4, label: "holdingRegister" },
] as const;

// Data type options
export const DATA_TYPES = [
  "Boolean",
  "Int16",
  "Int16BE",
  "Int32",
  "Int32BE",
  "UInt16",
  "UInt16BE",
  "UInt32",
  "UInt32BE",
  "Float32",
  "Float32BE",
  "Float64",
  "Float64BE",
  "String",
] as const;

// Policy options
export const POLICIES = [
  { value: "save", label: "policySave" },
  { value: "nosave", label: "policyNoSave" },
  { value: "onchange", label: "policyOnChange" },
] as const;

// Error view options
export const ERROR_VIEWS = [
  { value: "show", label: "showError" },
  { value: "hide", label: "hideError" },
  { value: "last", label: "showLastValue" },
] as const;

// Driver options
export const DRIVERS = [
  "modbus",
  "opcua",
  "bacnet",
  "mqtt",
  "s7",
  "fins",
] as const;
