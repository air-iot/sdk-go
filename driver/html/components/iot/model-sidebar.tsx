"use client";

import { useState } from "react";
import {
  ChevronRight,
  ChevronDown,
  Plus,
  MoreVertical,
  Server,
  Cpu,
  Search,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useI18n } from "@/lib/i18n";
import type { Model, Device } from "@/lib/iot-types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";

interface ModelSidebarProps {
  models: Model[];
  selectedModel: Model | null;
  selectedDevice: Device | null;
  onSelectModel: (model: Model) => void;
  onSelectDevice: (model: Model, device: Device) => void;
  onAddModel: () => void;
  onEditModel: (model: Model) => void;
  onDeleteModel: (model: Model) => void;
  onAddDevice: (model: Model) => void;
  onEditDevice: (model: Model, device: Device) => void;
  onDeleteDevice: (model: Model, device: Device) => void;
}

export function ModelSidebar({
  models,
  selectedModel,
  selectedDevice,
  onSelectModel,
  onSelectDevice,
  onAddModel,
  onEditModel,
  onDeleteModel,
  onAddDevice,
  onEditDevice,
  onDeleteDevice,
}: ModelSidebarProps) {
  const { t } = useI18n();
  const [expandedModels, setExpandedModels] = useState<Set<string>>(new Set());
  const [searchQuery, setSearchQuery] = useState("");

  const toggleExpand = (modelId: string) => {
    const newExpanded = new Set(expandedModels);
    if (newExpanded.has(modelId)) {
      newExpanded.delete(modelId);
    } else {
      newExpanded.add(modelId);
    }
    setExpandedModels(newExpanded);
  };

  const filteredModels = models.filter((model) => {
    const matchesModel = model.id.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesDevice = model.devices.some(
      (device) =>
        device.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
        device.name.toLowerCase().includes(searchQuery.toLowerCase())
    );
    return matchesModel || matchesDevice;
  });

  return (
    <div className="flex h-full w-72 flex-col border-r border-border bg-sidebar">
      <div className="flex items-center justify-between border-b border-border p-4">
        <h2 className="text-sm font-semibold text-sidebar-foreground">
          {t("modelList")}
        </h2>
        <Button
          size="sm"
          variant="ghost"
          onClick={onAddModel}
          className="h-8 w-8 p-0"
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>

      <div className="border-b border-border p-3">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder={t("search")}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-9 bg-sidebar-accent pl-9 text-sm"
          />
        </div>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-2">
          {filteredModels.length === 0 ? (
            <div className="py-8 text-center text-sm text-muted-foreground">
              {t("noModels")}
            </div>
          ) : (
            filteredModels.map((model) => (
              <div key={model.id} className="mb-1">
                <div
                  className={cn(
                    "group flex items-center rounded-md px-2 py-2 text-sm transition-colors",
                    selectedModel?.id === model.id && !selectedDevice
                      ? "bg-sidebar-accent text-sidebar-accent-foreground"
                      : "text-sidebar-foreground hover:bg-sidebar-accent/50"
                  )}
                >
                  <button
                    onClick={() => toggleExpand(model.id)}
                    className="mr-1 rounded p-0.5 hover:bg-sidebar-accent"
                  >
                    {expandedModels.has(model.id) ? (
                      <ChevronDown className="h-4 w-4" />
                    ) : (
                      <ChevronRight className="h-4 w-4" />
                    )}
                  </button>
                  <button
                    onClick={() => onSelectModel(model)}
                    className="flex flex-1 items-center gap-2"
                  >
                    <Server className="h-4 w-4 text-primary" />
                    <span className="flex-1 truncate text-left font-medium">
                      {model.id}
                    </span>
                    <Badge variant="secondary" className="ml-auto text-xs">
                      {model.devices.length}
                    </Badge>
                  </button>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="ml-1 h-6 w-6 p-0 opacity-0 group-hover:opacity-100"
                      >
                        <MoreVertical className="h-3 w-3" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => onAddDevice(model)}>
                        {t("addDevice")}
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => onEditModel(model)}>
                        {t("editModel")}
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        onClick={() => onDeleteModel(model)}
                        className="text-destructive"
                      >
                        {t("deleteModel")}
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>

                {expandedModels.has(model.id) && (
                  <div className="ml-6 mt-1 space-y-0.5">
                    {model.devices.length === 0 ? (
                      <div className="py-2 pl-4 text-xs text-muted-foreground">
                        {t("noDevices")}
                      </div>
                    ) : (
                      model.devices.map((device) => (
                        <div
                          key={device.id}
                          className={cn(
                            "group flex items-center rounded-md px-2 py-1.5 text-sm transition-colors",
                            selectedDevice?.id === device.id
                              ? "bg-primary/10 text-primary"
                              : "text-sidebar-foreground hover:bg-sidebar-accent/50"
                          )}
                        >
                          <button
                            onClick={() => onSelectDevice(model, device)}
                            className="flex flex-1 items-center gap-2"
                          >
                            <Cpu className="h-3.5 w-3.5" />
                            <span className="flex-1 truncate text-left">
                              {device.name || device.id}
                            </span>
                            <span
                              className={cn(
                                "h-2 w-2 rounded-full",
                                device.off
                                  ? "bg-muted-foreground"
                                  : device.disable
                                    ? "bg-warning"
                                    : "bg-primary"
                              )}
                            />
                          </button>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button
                                variant="ghost"
                                size="sm"
                                className="ml-1 h-5 w-5 p-0 opacity-0 group-hover:opacity-100"
                              >
                                <MoreVertical className="h-3 w-3" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem
                                onClick={() => onEditDevice(model, device)}
                              >
                                {t("editDevice")}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => onDeleteDevice(model, device)}
                                className="text-destructive"
                              >
                                {t("deleteDevice")}
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                      ))
                    )}
                  </div>
                )}
              </div>
            ))
          )}
        </div>
      </ScrollArea>
    </div>
  );
}
