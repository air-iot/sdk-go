package app

var schema = `{
    "title": "加法",
    "input": {
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
      "required": [
        "num1",
        "num2"
      ]
    },
    "output": {
      "type": "object",
      "properties": {
        "num1": {
          "title": "结果",
          "type": "number"
        }
      }
    }
  }`
