"use client";

import { I18nProvider } from "@/lib/i18n";
import { IotConfig } from "@/components/iot/iot-config";
import { Toaster } from "@/components/ui/sonner";

export default function Home() {
  return (
    <I18nProvider>
      <IotConfig />
      <Toaster />
    </I18nProvider>
  );
}
