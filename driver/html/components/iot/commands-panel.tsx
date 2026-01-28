"use client";

import { useState } from "react";
import { Plus, Pencil, Trash2, Copy, ChevronDown, ChevronRight } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import type { ModelConfig, DriverSchema, CommandConfig } from "@/lib/api";
import type { FormField, Operation } from "@/lib/iot-types";
import { DATA_TYPES } from "@/lib/iot-types";
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
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { JsonSchemaForm } from "./json-schema-form";
import { cn } from "@/lib/utils";

interface CommandsPanelProps {
  model: ModelConfig;
  schema: DriverSchema | null;
  onConfigChange: (updatedModel: ModelConfig) => void;
  onRunCommand?: (commandName: string, params?: Record<string, unknown>) => void;
}

const defaultCommand: any = {
  name: "",
};

export function CommandsPanel({
  model,
  schema,
  onConfigChange,
  onRunCommand,
}: CommandsPanelProps) {
  const { t } = useI18n();
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [expandedCommands, setExpandedCommands] = useState<Set<number>>(new Set());
  const [currentCommand, setCurrentCommand] = useState<any>(defaultCommand);

  // 获取commands列表（兼容新旧数据格式）
  const commands = (model.device?.commands || []) as any[];

  // 获取 command 的 schema (兼容多种可能的路径)
  const commandSchema = (schema?.driver?.properties?.commands as any)?.items ||
    (schema?.model?.properties?.commands as any)?.items ||
    null;

  const handleAddCommand = (command: any) => {
    onConfigChange({
      ...model,
      device: {
        ...model.device,
        commands: [...commands, command],
      },
    });
  };

  const handleUpdateCommand = (index: number, command: any) => {
    const newCommands = [...commands];
    newCommands[index] = command;
    onConfigChange({
      ...model,
      device: {
        ...model.device,
        commands: newCommands,
      },
    });
  };

  const handleDeleteCommand = (index: number) => {
    const newCommands = commands.filter((_, i) => i !== index);
    onConfigChange({
      ...model,
      device: {
        ...model.device,
        commands: newCommands,
      },
    });
  };

  const toggleExpand = (index: number) => {
    const newExpanded = new Set(expandedCommands);
    if (newExpanded.has(index)) {
      newExpanded.delete(index);
    } else {
      newExpanded.add(index);
    }
    setExpandedCommands(newExpanded);
  };

  const handleAdd = () => {
    setEditingIndex(null);
    setCurrentCommand({} as any);
    setIsDialogOpen(true);
  };

  const handleEdit = (index: number) => {
    setEditingIndex(index);
    setCurrentCommand(JSON.parse(JSON.stringify(commands[index])));
    setIsDialogOpen(true);
  };

  const handleDuplicate = (index: number) => {
    const command = JSON.parse(JSON.stringify(commands[index]));
    command.name = `${command.name}_copy`;
    handleAddCommand(command);
  };

  const handleSave = () => {
    if (editingIndex !== null) {
      handleUpdateCommand(editingIndex, currentCommand);
    } else {
      handleAddCommand(currentCommand);
    }
    setIsDialogOpen(false);
  };

  const handleRunCommand = (commandName: string) => {
    if (onRunCommand) {
      const command = commands.find(c => c.name === commandName);
      if (command) {
        onRunCommand(commandName, command.defaultValue);
      }
    }
  };

  return (
    <Card className="border-border bg-card">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            {t("commandList")}
            <Badge variant="secondary">{commands.length}</Badge>
          </CardTitle>
          <Button size="sm" onClick={handleAdd}>
            <Plus className="mr-1.5 h-4 w-4" />
            {t("addCommand")}
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[400px]">
          {commands.length === 0 ? (
            <div className="flex h-32 items-center justify-center text-muted-foreground">
              {t("noCommands")}
            </div>
          ) : (
            <div className="space-y-2">
              {commands.map((command, index) => (
                <Collapsible
                  key={index}
                  open={expandedCommands.has(index)}
                  onOpenChange={() => toggleExpand(index)}
                >
                  <div className="rounded-lg border border-border bg-secondary/20">
                    <CollapsibleTrigger asChild>
                      <div className="flex cursor-pointer items-center justify-between p-3 hover:bg-secondary/40">
                        <div className="flex items-center gap-3">
                          {expandedCommands.has(index) ? (
                            <ChevronDown className="h-4 w-4" />
                          ) : (
                            <ChevronRight className="h-4 w-4" />
                          )}
                          <span className="font-medium">{command.name || `Command ${index}`}</span>
                          <Badge variant="outline" className="text-xs">
                            {Object.keys(command).length} fields
                          </Badge>
                        </div>
                        <div className="flex items-center gap-1">
                          {onRunCommand && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-7 px-2 text-xs"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleRunCommand(command.name);
                              }}
                            >
                              Run
                            </Button>
                          )}
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-7 w-7 p-0"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDuplicate(index);
                            }}
                          >
                            <Copy className="h-3.5 w-3.5" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-7 w-7 p-0"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleEdit(index);
                            }}
                          >
                            <Pencil className="h-3.5 w-3.5" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-7 w-7 p-0 text-destructive hover:text-destructive"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDeleteCommand(index);
                            }}
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      </div>
                    </CollapsibleTrigger>
                    <CollapsibleContent>
                      <div className="border-t border-border p-3 text-sm">
                        <pre className="font-mono text-xs overflow-auto max-h-40 bg-secondary/10 p-2 rounded">
                          {JSON.stringify(command, null, 2)}
                        </pre>
                      </div>
                    </CollapsibleContent>
                  </div>
                </Collapsible>
              ))}
            </div>
          )}
        </ScrollArea>
      </CardContent>

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {editingIndex !== null ? t("editCommand") : t("addCommand")}
            </DialogTitle>
          </DialogHeader>

          <div className="py-4">
            {commandSchema ? (
              <JsonSchemaForm
                schema={commandSchema}
                value={currentCommand as any}
                onChange={(values) => setCurrentCommand(values as any)}
              />
            ) : (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="commandName">{t("commandName")}</Label>
                  <Input
                    id="commandName"
                    value={currentCommand.name || ""}
                    onChange={(e) =>
                      setCurrentCommand({ ...currentCommand, name: e.target.value })
                    }
                    className="bg-input"
                  />
                </div>
                <p className="text-sm text-muted-foreground">
                  No command schema available. Only basic name can be edited.
                </p>
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
              {t("cancel")}
            </Button>
            <Button onClick={handleSave}>{t("save")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
