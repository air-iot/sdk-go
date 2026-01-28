"use client";

import { useState } from "react";
import { Plus, Pencil, Trash2, Copy, Search } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import type { ModelConfig, DriverSchema, TagConfig } from "@/lib/api";
import type { Tag } from "@/lib/iot-types";
import { MODBUS_AREAS, DATA_TYPES, POLICIES, ERROR_VIEWS } from "@/lib/iot-types";
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
import { JsonSchemaForm } from "./json-schema-form";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface TagsPanelProps {
  model: ModelConfig;
  schema: DriverSchema | null;
  onConfigChange: (updatedModel: ModelConfig) => void;
}

const defaultTag: Tag = {
  id: "",
  name: "",
};

export function TagsPanel({
  model,
  schema,
  onConfigChange,
}: TagsPanelProps) {
  const { t } = useI18n();
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [currentTag, setCurrentTag] = useState<Record<string, unknown>>({});
  const [searchQuery, setSearchQuery] = useState("");
  const [activeTab, setActiveTab] = useState<"basic" | "advanced">("basic");

  // 获取tags列表（兼容新旧数据格式）
  const tags = (model.device?.tags || []) as Tag[];

  const filteredTags = tags.filter(
    (tag) =>
      tag.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
      tag.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // 获取 tag 的 schema (兼容多种可能的路径)
  const tagSchema = (schema?.driver?.properties?.tags as any)?.items ||
    (schema?.model?.properties?.tags as any)?.items ||
    null;

  const handleAddTag = (tag: Tag) => {
    onConfigChange({
      ...model,
      device: {
        ...model.device,
        tags: [...tags, tag],
      },
    });
  };

  const handleUpdateTag = (index: number, tag: Tag) => {
    const newTags = [...tags];
    newTags[index] = tag;
    onConfigChange({
      ...model,
      device: {
        ...model.device,
        tags: newTags,
      },
    });
  };

  const handleDeleteTag = (tag: Tag) => {
    const newTags = tags.filter((t) => t.id !== tag.id);
    onConfigChange({
      ...model,
      device: {
        ...model.device,
        tags: newTags,
      },
    });
  };

  const handleAdd = () => {
    setEditingIndex(null);
    setCurrentTag({ ...defaultTag } as any);
    setActiveTab("basic");
    setIsDialogOpen(true);
  };

  const handleEdit = (tag: Tag) => {
    const index = tags.findIndex((t) => t.id === tag.id);
    setEditingIndex(index);
    setCurrentTag({ ...tag } as any);
    setActiveTab("basic");
    setIsDialogOpen(true);
  };

  const handleDuplicate = (tag: Tag) => {
    const newTag = { ...tag, id: `${tag.id}_copy` };
    onConfigChange({
      ...model,
      device: {
        ...model.device,
        tags: [...tags, newTag],
      },
    });
  };

  const handleSave = () => {
    const tagToSave = editingIndex !== null
      ? { ...tags[editingIndex], ...currentTag }
      : (currentTag as any);

    if (editingIndex !== null) {
      const newTags = [...tags];
      newTags[editingIndex] = tagToSave;
      onConfigChange({
        ...model,
        device: {
          ...model.device,
          tags: newTags,
        },
      });
    } else {
      onConfigChange({
        ...model,
        device: {
          ...model.device,
          tags: [...tags, tagToSave as Tag],
        },
      });
    }
    setIsDialogOpen(false);
  };

  const getAreaLabel = (area: number) => {
    const found = MODBUS_AREAS.find((a) => a.value === area);
    return found ? t(found.label as any) : area;
  };

  return (
    <Card className="border-border bg-card">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            {t("tagList")}
            <Badge variant="secondary">{tags.length}</Badge>
          </CardTitle>
          <Button size="sm" onClick={handleAdd}>
            <Plus className="mr-1.5 h-4 w-4" />
            {t("addTag")}
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="mb-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder={t("search")}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="bg-input pl-9"
            />
          </div>
        </div>

        <ScrollArea className="h-[400px]">
          {filteredTags.length === 0 ? (
            <div className="flex h-32 items-center justify-center text-muted-foreground">
              {t("noTags")}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow className="border-border hover:bg-transparent">
                  <TableHead className="w-[150px]">{t("tagId")}</TableHead>
                  <TableHead>{t("tagName")}</TableHead>
                  <TableHead className="text-right">
                    {t("action")}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredTags.map((tag) => (
                  <TableRow key={tag.id} className="border-border">
                    <TableCell className="font-mono text-sm">
                      {tag.id}
                    </TableCell>
                    <TableCell>{tag.name}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 w-7 p-0"
                          onClick={() => handleDuplicate(tag)}
                        >
                          <Copy className="h-3.5 w-3.5" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 w-7 p-0"
                          onClick={() => handleEdit(tag)}
                        >
                          <Pencil className="h-3.5 w-3.5" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 w-7 p-0 text-destructive hover:text-destructive"
                          onClick={() => handleDeleteTag(tag)}
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
        <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {editingIndex !== null ? t("editTag") : t("addTag")}
            </DialogTitle>
          </DialogHeader>

          <div className="grid gap-4 py-4 max-h-[60vh] overflow-y-auto pr-2">
            <JsonSchemaForm
              schema={tagSchema || {
                type: "object",
                properties: {
                  id: { type: "string", title: t("tagId") },
                  name: { type: "string", title: t("tagName") }
                }
              }}
              value={currentTag}
              onChange={(values) => setCurrentTag(values)}
            />
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
