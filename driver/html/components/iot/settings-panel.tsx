"use client";

import { useI18n } from "@/lib/i18n";
import type { ModelConfig, DriverSchema } from "@/lib/api";
import { JsonSchemaForm } from "./json-schema-form";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Network, Settings2 } from "lucide-react";

interface SettingsPanelProps {
  model: ModelConfig;
  schema: DriverSchema | null;
  onConfigChange: (updatedModel: ModelConfig) => void;
}

export function SettingsPanel({
  model,
  schema,
  onConfigChange,
}: SettingsPanelProps) {
  const { t } = useI18n();

  // 获取模型级别的设置schema (对应 schema.go 中的 model.properties.settings 或 device.properties.settings)
  const modelSchema = schema?.model?.properties?.settings ||
    schema?.device?.properties?.settings ||
    null;

  // 处理模型设置更新
  const handleModelSettingsChange = (newSettings: Record<string, unknown>) => {
    onConfigChange({
      ...model,
      device: {
        ...model.device,
        settings: {
          ...model.device.settings,
          ...newSettings,
        },
      },
    });
  };

  // 获取当前模型设置值
  const modelSettingsValue = (model.device.settings || {}) as Record<string, unknown>;

  return (
    <div className="space-y-6">
      {/* Schema驱动的模型设置表单 */}
      {modelSchema ? (
        <JsonSchemaForm
          schema={modelSchema}
          value={modelSettingsValue}
          onChange={handleModelSettingsChange}
          title={modelSchema.title || t("connectionSettings") || "模型配置"}
        />
      ) : (
        /* 无schema时的后备显示 */
        <Card className="border-border bg-card">
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <Network className="h-4 w-4 text-primary" />
              {t("connectionSettings")}
              <Badge variant="outline" className="ml-auto">
                {model.device.driver.toUpperCase()}
              </Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              No settings schema available for this model.
            </p>
          </CardContent>
        </Card>
      )}

      {/* 设备级别显示（针对第一个设备） */}
      {model.devices && model.devices.length > 0 && (
        <Card className="border-border bg-card">
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base">
              <Settings2 className="h-4 w-4 text-primary" />
              Device: {model.devices[0].name || model.devices[0].id}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="text-sm text-muted-foreground">
              Device ID: {model.devices[0].id}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
