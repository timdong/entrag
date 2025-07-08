#!/bin/bash

echo "预加载Ollama模型到内存..."

# 下载并预加载更快的模型
echo "下载llama3.2:3b模型..."
ollama pull llama3.2:3b

echo "预加载模型到内存..."
curl -X POST http://localhost:11434/api/generate -d '{
  "model": "llama3.2:3b",
  "prompt": "Hello",
  "stream": false
}'

echo "模型预加载完成！"

# 检查当前加载的模型
echo "当前内存中的模型："
curl -X GET http://localhost:11434/api/ps 