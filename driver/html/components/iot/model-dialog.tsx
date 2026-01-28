"use client";

import { useState, useEffect } from "react";
import { useI18n } from "@/lib/i18n";
import type { ModelConfig, DriverSchema } from "@/lib/api";
import { DRIVERS } from "@/lib/iot-types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { JsonSchemaForm } from "./json-schema-form";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";

interface ModelDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  model: ModelConfig | null;
  schema: DriverSchema | null;
  onSave: (model: ModelConfig) => void;
}

const defaultModel: ModelConfig = {
  id: "",
  name: "",
  device: {
    driver: "modbus",
    groupId: "",
    settings: {},
    tags: [],
    commands: [],
    events: [],
  },
  devices: [],
};

export function ModelDialog({
  open,
  onOpenChange,
  model,
  schema,
  onSave,
}: ModelDialogProps) {
  const { t } = useI18n();
  const [currentModel, setCurrentModel] = useState<ModelConfig>(defaultModel);
  const [activeTab, setActiveTab] = useState<"basic" | "settings">("basic");

  useEffect(() => {
    if (model) {
      setCurrentModel(JSON.parse(JSON.stringify(model)));
    } else {
      setCurrentModel(JSON.parse(JSON.stringify(defaultModel)));
    }
  }, [model, open]);

  const handleSave = () => {
    onSave(currentModel);
    onOpenChange(false);
  };

  // 获取设置部分的 schema (对应 schema.go 中的 model.properties.settings 或 device.properties.settings)
  const settingsSchema = schema?.model?.properties?.settings ||
    schema?.device?.properties?.settings ||
    null;

  const handleSettingsChange = (newSettings: Record<string, unknown>) => {
    setCurrentModel({
      ...currentModel,
      device: {
        ...currentModel.device,
        settings: {
          ...currentModel.device.settings,
          ...newSettings,
        },
      },
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {model ? t("editModel") : t("addModel")}
          </DialogTitle>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as "basic" | "settings")} className="w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="basic">基本信息</TabsTrigger>
            <TabsTrigger value="settings">连接设置</TabsTrigger>
          </TabsList>

          <TabsContent value="basic" className="space-y-4 mt-4">
            <div className="space-y-2">
              <Label htmlFor="modelId">模型 ID *</Label>
              <Input
                id="modelId"
                value={currentModel.id}
                onChange={(e) =>
                  setCurrentModel({ ...currentModel, id: e.target.value })
                }
                className="bg-input"
                disabled={!!model}
                placeholder="例如: model_001"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="modelName">模型名称</Label>
              <Input
                id="modelName"
                value={currentModel.name}
                onChange={(e) =>
                  setCurrentModel({ ...currentModel, name: e.target.value })
                }
                className="bg-input"
                placeholder="例如: Modbus TCP 设备组"
              />
            </div>
          </TabsContent>

          <TabsContent value="settings" className="mt-4">
            {settingsSchema ? (
              <ScrollArea className="h-[400px] pr-4">
                <JsonSchemaForm
                  schema={settingsSchema}
                  value={currentModel.device.settings as Record<string, unknown>}
                  onChange={handleSettingsChange}
                />
              </ScrollArea>
            ) : (
              <div className="text-center py-8 text-sm text-muted-foreground">
                暂无配置 Schema
                <p className="mt-2 text-xs">
                  驱动未提供设置 Schema，请手动配置
                </p>
              </div>
            )}
          </TabsContent>
        </Tabs>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("cancel")}
          </Button>
          <Button onClick={handleSave} disabled={!currentModel.id}>
            {t("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
