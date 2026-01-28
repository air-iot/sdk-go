"use client"

import { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog"
import { Header } from "./header"
import { ModelsPanel } from "./models-panel"
import { DevicesPanel } from "./devices-panel"
import { TagsPanel } from "./tags-panel"
import { CommandsPanel } from "./commands-panel"
import { EventsPanel } from "./events-panel"
import { SettingsPanel } from "./settings-panel"
import { JsonSchemaForm } from "./json-schema-form"
import { getConfig, saveConfig, getSchema, getInfo, type DriverConfig, type DriverSchema, runCommand } from "@/lib/api"
import { toast } from "@/hooks/use-toast"

export function IotConfig() {
  const [config, setConfig] = useState<DriverConfig | null>(null)
  const [schema, setSchema] = useState<DriverSchema | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [selectedModelId, setSelectedModelId] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'devices' | 'tags' | 'commands' | 'events' | 'settings'>('devices')

  // 加载配置和 schema
  useEffect(() => {
    async function loadData() {
      setLoading(true)
      try {
        const configData = await getConfig()
        setConfig(configData)

        if (configData && configData.tables?.length > 0) {
          setSelectedModelId(configData.tables[0].id)
        }

        const schemaData = await getSchema()
        setSchema(schemaData)
      } catch (error) {
        toast({
          title: "加载失败",
          description: "无法加载配置信息",
          variant: "destructive"
        })
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [])

  // 保存配置
  async function handleSave() {
    if (!config) return
    setSaving(true)
    try {
      const result = await saveConfig(config)
      if (result.success) {
        toast({
          title: "保存成功",
          description: result.message
        })
      } else {
        throw new Error(result.message)
      }
    } catch (error) {
      toast({
        title: "保存失败",
        description: error instanceof Error ? error.message : "无法保存配置信息",
        variant: "destructive"
      })
    } finally {
      setSaving(false)
    }
  }

  // 获取当前选中的模型
  const selectedModel = config?.tables?.find(m => m.id === selectedModelId)

  // 执行指令
  async function handleRunCommand(commandName: string, params?: Record<string, unknown>) {
    if (!config || !selectedModelId) return

    const device = selectedModel?.devices?.[0]
    if (!device) {
      toast({
        title: "执行失败",
        description: "请先添加设备",
        variant: "destructive"
      })
      return
    }

    const result = await runCommand(selectedModelId, device.id, commandName, params)
    if (result.success) {
      toast({
        title: "执行成功",
        description: result.message
      })
    } else {
      toast({
        title: "执行失败",
        description: result.message,
        variant: "destructive"
      })
    }
  }

  const [isDriverSettingsOpen, setIsDriverSettingsOpen] = useState(false)

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-lg">加载中...</div>
      </div>
    )
  }

  // 如果配置为空，初始化一个空配置结构
  const effectiveConfig = config || {
    id: "",
    name: "新驱动配置",
    groupId: "",
    driverType: "modbus",
    runMode: "local",
    device: {
      settings: {},
      tags: [],
      commands: [],
      events: []
    },
    distributed: "",
    ports: "",
    tables: []
  }

  return (
    <div className="flex flex-col min-h-screen">
      <Header
        config={effectiveConfig}
        onSave={handleSave}
        saving={saving}
        onOpenDriverSettings={() => setIsDriverSettingsOpen(true)}
        onExport={() => {/* TODO: Implement export */ }}
        onImport={() => {/* TODO: Implement import */ }}
      />
      <Dialog open={isDriverSettingsOpen} onOpenChange={setIsDriverSettingsOpen}>
        <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>驱动实例配置</DialogTitle>
          </DialogHeader>
          <div className="py-4">
            {schema?.driver?.properties?.settings ? (
              <JsonSchemaForm
                schema={schema.driver.properties.settings}
                value={effectiveConfig.device?.settings || {}}
                onChange={(values) => {
                  setConfig({
                    ...effectiveConfig,
                    device: {
                      ...effectiveConfig.device,
                      settings: values
                    }
                  })
                }}
              />
            ) : (
              <div className="text-center py-8 text-sm text-muted-foreground">
                暂无驱动实例配置 Schema
              </div>
            )}
          </div>
          <DialogFooter>
            <Button onClick={() => setIsDriverSettingsOpen(false)}>确定</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <div className="flex flex-1">
        <ModelsPanel
          config={effectiveConfig}
          selectedModelId={selectedModelId}
          onSelectModel={setSelectedModelId}
          onConfigChange={setConfig}
          schema={schema}
        />
        {selectedModel && (
          <div className="flex-1 flex flex-col overflow-hidden">
            <div className="border-b">
              <div className="flex">
                <button
                  className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'devices'
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                    }`}
                  onClick={() => setActiveTab('devices')}
                >
                  设备
                </button>
                <button
                  className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'tags'
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                    }`}
                  onClick={() => setActiveTab('tags')}
                >
                  点位
                </button>
                <button
                  className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'commands'
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                    }`}
                  onClick={() => setActiveTab('commands')}
                >
                  指令
                </button>
                <button
                  className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'events'
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                    }`}
                  onClick={() => setActiveTab('events')}
                >
                  事件
                </button>
                <button
                  className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${activeTab === 'settings'
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                    }`}
                  onClick={() => setActiveTab('settings')}
                >
                  设置
                </button>
              </div>
            </div>
            <div className="flex-1 overflow-auto p-4">
              {activeTab === 'devices' && (
                <DevicesPanel
                  model={selectedModel}
                  schema={schema}
                  onConfigChange={(updatedModel) => {
                    setConfig({
                      ...effectiveConfig,
                      tables: effectiveConfig.tables.map(m =>
                        m.id === updatedModel.id ? updatedModel : m
                      )
                    })
                  }}
                />
              )}
              {activeTab === 'tags' && (
                <TagsPanel
                  model={selectedModel}
                  schema={schema}
                  onConfigChange={(updatedModel) => {
                    setConfig({
                      ...effectiveConfig,
                      tables: effectiveConfig.tables.map(m =>
                        m.id === updatedModel.id ? updatedModel : m
                      )
                    })
                  }}
                />
              )}
              {activeTab === 'commands' && (
                <CommandsPanel
                  model={selectedModel}
                  schema={schema}
                  onConfigChange={(updatedModel) => {
                    setConfig({
                      ...effectiveConfig,
                      tables: effectiveConfig.tables.map(m =>
                        m.id === updatedModel.id ? updatedModel : m
                      )
                    })
                  }}
                  onRunCommand={handleRunCommand}
                />
              )}
              {activeTab === 'events' && (
                <EventsPanel
                  model={selectedModel}
                  schema={schema}
                  onConfigChange={(updatedModel) => {
                    setConfig({
                      ...effectiveConfig,
                      tables: effectiveConfig.tables.map(m =>
                        m.id === updatedModel.id ? updatedModel : m
                      )
                    })
                  }}
                />
              )}
              {activeTab === 'settings' && (
                <SettingsPanel
                  model={selectedModel}
                  schema={schema}
                  onConfigChange={(updatedModel) => {
                    setConfig({
                      ...effectiveConfig,
                      tables: effectiveConfig.tables.map(m =>
                        m.id === updatedModel.id ? updatedModel : m
                      )
                    })
                  }}
                />
              )}
            </div>
          </div>
        )}
        {!selectedModel && (
          <div className="flex-1 flex items-center justify-center bg-muted/20">
            <div className="text-center space-y-4">
              <p className="text-lg text-muted-foreground">
                {effectiveConfig.tables?.length === 0 ? "暂无模型配置" : "请从左侧选择一个模型"}
              </p>
              {effectiveConfig.tables?.length === 0 && (
                <p className="text-sm text-muted-foreground">
                  点击左侧面板的"添加模型"按钮创建第一个模型
                </p>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
