"use client";

import { useI18n, type Locale } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Download, Upload, Save, Activity, Globe, Settings2 } from "lucide-react";

interface HeaderProps {
  config: any;
  onSave: () => void;
  saving: boolean;
  onOpenDriverSettings: () => void;
  onExport: () => void;
  onImport: () => void;
}

export function Header({ config, onSave, saving, onOpenDriverSettings, onExport, onImport }: HeaderProps) {
  const { t, locale, setLocale } = useI18n();

  return (
    <header className="flex h-16 items-center justify-between border-b border-border bg-card px-6">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
          <Activity className="h-5 w-5 text-primary" />
        </div>
        <div>
          <h1 className="text-lg font-semibold text-foreground">{t("title")}</h1>
          <p className="text-xs text-muted-foreground">{t("subtitle")}</p>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <div className="flex items-center gap-2 rounded-lg border border-border bg-secondary/30 p-1">
          <Globe className="ml-2 h-4 w-4 text-muted-foreground" />
          <Select
            value={locale}
            onValueChange={(value) => setLocale(value as Locale)}
          >
            <SelectTrigger className="h-8 w-24 border-0 bg-transparent">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="zh">中文</SelectItem>
              <SelectItem value="en">English</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <Button variant="outline" size="sm" onClick={onImport}>
          <Upload className="mr-1.5 h-4 w-4" />
          {t("import")}
        </Button>

        <Button variant="outline" size="sm" onClick={onExport}>
          <Download className="mr-1.5 h-4 w-4" />
          {t("export")}
        </Button>

        <Button variant="outline" size="sm" onClick={onOpenDriverSettings}>
          <Settings2 className="mr-1.5 h-4 w-4" />
          驱动配置
        </Button>

        <Button size="sm" onClick={onSave} disabled={saving}>
          <Save className="mr-1.5 h-4 w-4" />
          {saving ? "保存中..." : t("save")}
        </Button>
      </div>
    </header>
  );
}
