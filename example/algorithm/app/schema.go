package app

var Schema = `[
  {
    "title": "函数1-加法",
    "function": "add",
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
  },
  {
    "title": "函数2-绝对值",
    "function": "abs",
    "input": {
      "type": "object",
      "properties": {
        "num1": {
          "title": "参数1",
          "type": "number"
        }
      },
      "required": [
        "num1"
      ]
    },
    "output": {
      "type": "object",
      "properties": {
        "res": {
          "title": "结果",
          "type": "number"
        }
      }
    }
  }
]`
