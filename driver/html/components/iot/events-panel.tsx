"use client";

import { useState } from "react";
import { Plus, Pencil, Trash2, Copy, Zap } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import type { ModelConfig, DriverSchema, EventConfig } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { JsonSchemaForm } from "./json-schema-form";
import { Switch } from "@/components/ui/switch";

interface EventsPanelProps {
  model: ModelConfig;
  schema: DriverSchema | null;
  onConfigChange: (updatedModel: ModelConfig) => void;
}

const defaultEvent: any = {
  name: "",
};

export function EventsPanel({
  model,
  schema,
  onConfigChange,
}: EventsPanelProps) {
  const { t } = useI18n();
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);

  // 获取events列表（兼容新旧数据格式）
  const events = (model.device?.events || []) as any[];

  // 获取 event 的 schema (兼容多种可能的路径)
  const eventSchema = (schema?.driver?.properties?.events as any)?.items ||
    (schema?.model?.properties?.events as any)?.items ||
    null;

  const [currentEvent, setCurrentEvent] = useState<any>({});

  const handleAddEvent = (event: any) => {
    onConfigChange({
      ...model,
      device: {
        ...model.device,
        events: [...events, event],
      },
    });
  };

  const handleUpdateEvent = (index: number, event: any) => {
    const newEvents = [...events];
    newEvents[index] = event;
    onConfigChange({
      ...model,
      device: {
        ...model.device,
        events: newEvents,
      },
    });
  };

  const handleDeleteEvent = (index: number) => {
    const newEvents = events.filter((_, i) => i !== index);
    onConfigChange({
      ...model,
      device: {
        ...model.device,
        events: newEvents,
      },
    });
  };

  const handleAdd = () => {
    setEditingIndex(null);
    setCurrentEvent({} as any);
    setIsDialogOpen(true);
  };

  const handleEdit = (index: number) => {
    setEditingIndex(index);
    setCurrentEvent({ ...events[index] } as any);
    setIsDialogOpen(true);
  };

  const handleDuplicate = (index: number) => {
    const event = { ...events[index], name: `${events[index].name}_copy`, id: `${events[index].id}_copy` };
    handleAddEvent(event);
  };

  const handleToggleEnabled = (index: number) => {
    const event = {
      ...events[index],
      enabled: !(events[index].enabled ?? events[index].enable),
      enable: !(events[index].enabled ?? events[index].enable)
    };
    handleUpdateEvent(index, event);
  };

  const handleSave = () => {
    const eventToSave = { ...currentEvent };
    if (editingIndex !== null) {
      handleUpdateEvent(editingIndex, eventToSave);
    } else {
      handleAddEvent(eventToSave);
    }
    setIsDialogOpen(false);
  };

  return (
    <Card className="border-border bg-card">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <Zap className="h-4 w-4 text-primary" />
            {t("eventList")}
            <Badge variant="secondary">{events.length}</Badge>
          </CardTitle>
          <Button size="sm" onClick={handleAdd}>
            <Plus className="mr-1.5 h-4 w-4" />
            {t("addEvent")}
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[400px]">
          {events.length === 0 ? (
            <div className="flex h-32 items-center justify-center text-muted-foreground">
              {t("noEvents")}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="border-border hover:bg-transparent">
                  <TableHead className="w-[150px]">{t("eventName")}</TableHead>
                  <TableHead className="w-[100px]">{t("enabled")}</TableHead>
                  <TableHead className="w-[100px] text-right">
                    {t("action")}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {events.map((event, index) => (
                  <TableRow key={event.id || index} className="border-border">
                    <TableCell className="font-medium">{event.name || `Event ${index}`}</TableCell>
                    <TableCell>
                      <Switch
                        checked={event.enabled ?? event.enable}
                        onCheckedChange={() => handleToggleEnabled(index)}
                      />
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 w-7 p-0"
                          onClick={() => handleDuplicate(index)}
                        >
                          <Copy className="h-3.5 w-3.5" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 w-7 p-0"
                          onClick={() => handleEdit(index)}
                        >
                          <Pencil className="h-3.5 w-3.5" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 w-7 p-0 text-destructive hover:text-destructive"
                          onClick={() => handleDeleteEvent(index)}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </ScrollArea>
      </CardContent>

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="sm:max-w-lg max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {editingIndex !== null ? t("editEvent") : t("addEvent")}
            </DialogTitle>
          </DialogHeader>

          <div className="py-4">
            {eventSchema ? (
              <JsonSchemaForm
                schema={eventSchema}
                value={currentEvent}
                onChange={(values) => setCurrentEvent(values)}
              />
            ) : (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="eventName">{t("eventName")}</Label>
                  <Input
                    id="eventName"
                    value={currentEvent.name || ""}
                    onChange={(e) =>
                      setCurrentEvent({ ...currentEvent, name: e.target.value })
                    }
                    className="bg-input"
                  />
                </div>
                <div className="flex items-center justify-between rounded-lg border border-border bg-secondary/30 p-3">
                  <Label className="text-sm">{t("enabled")}</Label>
                  <Switch
                    checked={currentEvent.enabled ?? currentEvent.enable}
                    onCheckedChange={(checked) =>
                      setCurrentEvent({ ...currentEvent, enabled: checked, enable: checked })
                    }
                  />
                </div>
                <p className="text-sm text-muted-foreground">
                  No event schema available. Only basic name and status can be edited.
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
