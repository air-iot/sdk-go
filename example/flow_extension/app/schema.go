package app

var schema = `{
	"type": "object",
	"properties": {
		"num1": {
			"title": "参数1",
			"type": "number"
		},
		"num2": {
			"title": "参数2",
			"type": "number"
		}
	},
	"required": ["num1", "num2"]
}`
