package app

var Schema = `({
	"title": "测试SDK",
	"key": "test1",
	"driver": {
		"properties": {
			"settings": {
				"title": "实例配置",
				"type": "object",
				"properties": {
				},
				"required": []
			}
		}
	},
	"model": {
		"properties": {
			"settings": {
				"title": "模型配置",
				"type": "object",
				"properties": {
				},
				"required": []
			},
			"tags": {
				"title": "数据点",
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"id": {
							"type": "string",
							"title": "标识"
						},
						"name": {
							"type": "string",
							"title": "名称"
						},

						"unit": {
							"type": "string",
							"title": "单位"
						}
					},
					"required": [
						"id", "name"
					]
				}
			},
			"commands": {
				"title": "命令",
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"name": {
							"type": "string",
							"title": "name"
						},
						"form": {
							"type": "array",
							"title": "表单项",
							"items": {

							}
						},
						"ops": {
							"title": "指令",
							"type": "array",
							"items": {
								"type": "object",
								"properties": {}
							}
						}
					},
					"required": [
						"value"
					]
				}
			}
		}
	},
	"device": {
		"properties": {
			"settings": {
				"title": "设备配置",
				"type": "object",
				"properties": {}
			},
			"tags": {
				"title": "数据点",
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"name": {
							"type": "string",
							"title": "名称"
						},
						"id": {
							"type": "string",
							"title": "标识"
						}
					},
					"required": [
						"id", "name"
					]
				}
			},
			"commands": {
				"title": "命令",
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"name": {
							"type": "string",
							"title": "name"
						},
						"form": {
							"type": "array",
							"title": "表单项",
							"items": {

							}
						},
						"ops": {
							"title": "指令",
							"type": "array",
							"items": {
								"type": "object",
								"properties": {}
							}
						}
					},
					"required": [
					]
				}
			}
		}
	}
})`
