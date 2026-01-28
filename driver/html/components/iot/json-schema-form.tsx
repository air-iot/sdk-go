"use client";

import { useState } from "react";
import { useI18n } from "@/lib/i18n";
import type { JSONSchema, JSONSchemaProperty } from "@/lib/iot-types";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Plus, Trash2, FileJson } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Code } from "lucide-react";

interface JsonSchemaFormProps {
  schema: JSONSchemaProperty;
  value: Record<string, unknown>;
  onChange: (value: Record<string, unknown>) => void;
  title?: string;
}

function SchemaField({
  name,
  property,
  value,
  onChange,
}: {
  name: string;
  property: JSONSchemaProperty;
  value: unknown;
  onChange: (value: unknown) => void;
}) {
  const { t } = useI18n();

  switch (property.type) {
    case "boolean":
      return (
        <div className="flex items-center justify-between rounded-lg border border-border bg-secondary/20 p-3">
          <div>
            <Label className="text-sm font-medium">
              {property.title || name}
            </Label>
            {property.description && (
              <p className="text-xs text-muted-foreground">
                {property.description}
              </p>
            )}
          </div>
          <Switch
            checked={Boolean(value)}
            onCheckedChange={(checked) => onChange(checked)}
          />
        </div>
      );

    case "number":
    case "integer":
      return (
        <div className="space-y-2">
          <Label className="text-sm text-muted-foreground">
            {property.title || name}
          </Label>
          <Input
            type="number"
            value={value as number ?? property.default ?? 0}
            onChange={(e) => onChange(parseFloat(e.target.value) || 0)}
            min={property.minimum}
            max={property.maximum}
            className="bg-input"
          />
          {property.description && (
            <p className="text-xs text-muted-foreground">
              {property.description}
            </p>
          )}
        </div>
      );

    case "string":
      // 处理枚举类型（使用 enum_title 或 enumNames）
      if (property.enum) {
        return (
          <div className="space-y-2">
            <Label className="text-sm text-muted-foreground">
              {property.title || name}
            </Label>
            <Select
              value={String(value ?? property.default ?? "")}
              onValueChange={(v) => {
                if (property.type === "number" || property.type === "integer") {
                  onChange(parseFloat(v));
                } else {
                  onChange(v);
                }
              }}
            >
              <SelectTrigger className="bg-input">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {property.enum.map((option, index) => (
                  <SelectItem key={String(option)} value={String(option)}>
                    {property.enum_title?.[index] ?? property.enumNames?.[index] ?? String(option)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {property.description && (
              <p className="text-xs text-muted-foreground">
                {property.description}
              </p>
            )}
          </div>
        );
      }

      // 处理 fieldType
      if (property.fieldType === "password") {
        return (
          <div className="space-y-2">
            <Label className="text-sm text-muted-foreground">
              {property.title || name}
            </Label>
            <Input
              type="password"
              value={String(value ?? property.default ?? "")}
              onChange={(e) => onChange(e.target.value)}
              className="bg-input"
            />
            {property.description && (
              <p className="text-xs text-muted-foreground">
                {property.description}
              </p>
            )}
          </div>
        );
      }

      // 处理脚本编辑器类型
      if (property.fieldType === "deviceScriptEdit") {
        return (
          <div className="space-y-2">
            <Label className="text-sm text-muted-foreground flex items-center gap-2">
              <Code className="h-3.5 w-3.5" />
              {property.title || name}
            </Label>
            <Textarea
              value={String(value ?? property.defaultScript ?? property.default ?? "")}
              onChange={(e) => onChange(e.target.value)}
              className="bg-input font-mono text-xs min-h-[200px]"
              placeholder="请输入脚本代码..."
            />
            {property.description && (
              <p className="text-xs text-muted-foreground">
                {property.description}
              </p>
            )}
          </div>
        );
      }

      // 处理 textarea 格式
      if (property.format === "textarea" || property.fieldType === "textarea") {
        return (
          <div className="space-y-2">
            <Label className="text-sm text-muted-foreground">
              {property.title || name}
            </Label>
            <Textarea
              value={String(value ?? property.default ?? "")}
              onChange={(e) => onChange(e.target.value)}
              className="bg-input"
            />
            {property.description && (
              <p className="text-xs text-muted-foreground">
                {property.description}
              </p>
            )}
          </div>
        );
      }

      return (
        <div className="space-y-2">
          <Label className="text-sm text-muted-foreground">
            {property.title || name}
          </Label>
          <Input
            value={String(value ?? property.default ?? "")}
            onChange={(e) => onChange(e.target.value)}
            pattern={property.pattern}
            className="bg-input"
          />
          {property.description && (
            <p className="text-xs text-muted-foreground">
              {property.description}
            </p>
          )}
        </div>
      );

    case "array":
      const arrayValue = (value as unknown[]) || [];
      // 获取数组项的默认值
      const getArrayItemDefault = (): unknown => {
        if (property.items?.type === "object" && property.items?.properties) {
          // 对象类型：创建一个空对象，包含所有属性的默认值
          const defaultObj: Record<string, unknown> = {};
          Object.entries(property.items.properties).forEach(([key, prop]) => {
            if (prop.default !== undefined) {
              defaultObj[key] = prop.default;
            } else if (prop.type === "boolean") {
              defaultObj[key] = false;
            } else if (prop.type === "number" || prop.type === "integer") {
              defaultObj[key] = 0;
            } else if (prop.type === "array") {
              defaultObj[key] = [];
            } else if (prop.type === "object") {
              defaultObj[key] = {};
            } else {
              defaultObj[key] = "";
            }
          });
          return defaultObj;
        }
        return property.items?.default ?? "";
      };

      return (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <Label className="text-sm text-muted-foreground">
              {property.title || name}
            </Label>
            <Button
              size="sm"
              variant="outline"
              onClick={() =>
                onChange([...arrayValue, getArrayItemDefault()])
              }
            >
              <Plus className="mr-1 h-3 w-3" />
              添加
            </Button>
          </div>
          <div className="space-y-3">
            {arrayValue.length === 0 ? (
              <div className="text-center py-4 text-sm text-muted-foreground border border-dashed border-border rounded-lg">
                暂无项目，点击上方按钮添加
              </div>
            ) : (
              arrayValue.map((item, index) => (
                <div key={index} className="rounded-lg border border-border bg-secondary/20 p-3">
                  {property.items && (
                    <div className="flex gap-2">
                      <div className="flex-1">
                        <SchemaField
                          name={`${index}`}
                          property={property.items}
                          value={item}
                          onChange={(v) => {
                            const newArray = [...arrayValue];
                            newArray[index] = v;
                            onChange(newArray);
                          }}
                        />
                      </div>
                      <Button
                        size="sm"
                        variant="ghost"
                        className="h-8 w-8 p-0 text-destructive flex-shrink-0"
                        onClick={() => {
                          const newArray = arrayValue.filter((_, i) => i !== index);
                          onChange(newArray);
                        }}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        </div>
      );

    case "object":
      const objectValue = (value as Record<string, unknown>) || {};
      return (
        <div className="space-y-3 rounded-lg border border-border bg-secondary/10 p-3">
          <Label className="text-sm font-medium">
            {property.title || name}
          </Label>
          {property.properties &&
            Object.entries(property.properties).map(([key, prop]) => (
              <SchemaField
                key={key}
                name={key}
                property={prop}
                value={objectValue[key]}
                onChange={(v) => onChange({ ...objectValue, [key]: v })}
              />
            ))}
        </div>
      );

    default:
      return null;
  }
}

export function JsonSchemaForm({
  schema,
  value,
  onChange,
  title,
}: JsonSchemaFormProps) {
  const { t } = useI18n();

  const handleFieldChange = (name: string, fieldValue: unknown) => {
    onChange({ ...value, [name]: fieldValue });
  };

  return (
    <Card className="border-border bg-card">
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-base">
          <FileJson className="h-4 w-4 text-primary" />
          {title || schema.title || t("form")}
        </CardTitle>
        {schema.description && (
          <p className="text-sm text-muted-foreground">{schema.description}</p>
        )}
      </CardHeader>
      <CardContent className="space-y-4">
        {schema.properties && Object.entries(schema.properties).map(([name, property]) => (
          <SchemaField
            key={name}
            name={name}
            property={property}
            value={value[name]}
            onChange={(v) => handleFieldChange(name, v)}
          />
        ))}
      </CardContent>
    </Card>
  );
}

// Schema Editor Component for editing JSON schemas
interface SchemaEditorProps {
  schema: JSONSchema;
  onChange: (schema: JSONSchema) => void;
}

export function SchemaEditor({ schema, onChange }: SchemaEditorProps) {
  const { t } = useI18n();
  const [jsonText, setJsonText] = useState(JSON.stringify(schema, null, 2));
  const [error, setError] = useState<string | null>(null);

  const handleChange = (text: string) => {
    setJsonText(text);
    try {
      const parsed = JSON.parse(text);
      setError(null);
      onChange(parsed);
    } catch (e) {
      setError("Invalid JSON");
    }
  };

  return (
    <Card className="border-border bg-card">
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-base">
          <FileJson className="h-4 w-4 text-primary" />
          Schema Editor
          {error && (
            <Badge variant="destructive" className="ml-auto">
              {error}
            </Badge>
          )}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Textarea
          value={jsonText}
          onChange={(e) => handleChange(e.target.value)}
          className="h-64 bg-input font-mono text-xs"
        />
      </CardContent>
    </Card>
  );
}
