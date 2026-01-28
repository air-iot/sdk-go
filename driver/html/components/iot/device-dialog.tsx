"use client";

import { useState, useEffect } from "react";
import { useI18n } from "@/lib/i18n";
import type { Device } from "@/lib/iot-types";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Switch } from "@/components/ui/switch";
import { JsonSchemaForm } from "./json-schema-form";

interface DeviceDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  device: Device | null;
  schema: any; // Add schema prop
  onSave: (device: Device) => void;
}

const defaultDevice: Device = {
  id: "",
  name: "",
  device: {
    driver: "",
    groupId: "",
    settings: {},
    tags: [],
    commands: [],
    events: [],
  },
  disable: false,
  off: false,
};

export function DeviceDialog({
  open,
  onOpenChange,
  device,
  schema,
  onSave,
}: DeviceDialogProps) {
  const { t } = useI18n();
  const [currentDevice, setCurrentDevice] = useState<Device>(defaultDevice);
  const [activeTab, setActiveTab] = useState<"basic" | "settings">("basic");

  useEffect(() => {
    if (device) {
      setCurrentDevice(JSON.parse(JSON.stringify(device)));
    } else {
      setCurrentDevice(JSON.parse(JSON.stringify(defaultDevice)));
    }
  }, [device, open]);

  const handleSave = () => {
    onSave(currentDevice);
    onOpenChange(false);
  };

  // 获取设备级别的设置 schema (对应 schema.go 中的 device.properties.settings)
  const deviceSettingsSchema = schema?.device?.properties?.settings || null;

  const handleSettingsChange = (newSettings: Record<string, unknown>) => {
    setCurrentDevice({
      ...currentDevice,
      device: {
        ...currentDevice.device,
        settings: {
          ...currentDevice.device.settings,
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
            {device ? t("editDevice") : t("addDevice")}
          </DialogTitle>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as "basic" | "settings")} className="w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="basic">基本信息</TabsTrigger>
            <TabsTrigger value="settings">连接设置</TabsTrigger>
          </TabsList>

          <TabsContent value="basic" className="space-y-4 mt-4">
            <div className="space-y-2">
              <Label htmlFor="deviceId">{t("deviceId")}</Label>
              <Input
                id="deviceId"
                value={currentDevice.id}
                onChange={(e) =>
                  setCurrentDevice({ ...currentDevice, id: e.target.value })
                }
                className="bg-input"
                disabled={!!device}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="deviceName">{t("deviceName")}</Label>
              <Input
                id="deviceName"
                value={currentDevice.name}
                onChange={(e) =>
                  setCurrentDevice({ ...currentDevice, name: e.target.value })
                }
                className="bg-input"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="flex items-center justify-between rounded-lg border border-border bg-secondary/30 p-3">
                <Label className="text-sm">{t("enabled")}</Label>
                <Switch
                  checked={!currentDevice.disable}
                  onCheckedChange={(checked) =>
                    setCurrentDevice({ ...currentDevice, disable: !checked })
                  }
                />
              </div>
              <div className="flex items-center justify-between rounded-lg border border-border bg-secondary/30 p-3">
                <Label className="text-sm">{t("online")}</Label>
                <Switch
                  checked={!currentDevice.off}
                  onCheckedChange={(checked) =>
                    setCurrentDevice({ ...currentDevice, off: !checked })
                  }
                />
              </div>
            </div>
          </TabsContent>

          <TabsContent value="settings" className="mt-4">
            {deviceSettingsSchema ? (
              <ScrollArea className="h-[400px] pr-4">
                <JsonSchemaForm
                  schema={deviceSettingsSchema}
                  value={currentDevice.device.settings as Record<string, unknown>}
                  onChange={handleSettingsChange}
                />
              </ScrollArea>
            ) : (
              <div className="text-center py-8 text-sm text-muted-foreground">
                暂无设备配置 Schema
              </div>
            )}
          </TabsContent>
        </Tabs>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("cancel")}
          </Button>
          <Button
            onClick={handleSave}
            disabled={!currentDevice.id || !currentDevice.name}
          >
            {t("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
