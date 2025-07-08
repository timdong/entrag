#!/bin/bash

echo "🚀 构建优化版本的entrag..."

# 设置Go环境
export PATH=/usr/local/go/bin:$PATH

# 清理旧文件
echo "🧹 清理旧文件..."
rm -f entrag

# 构建优化版本
echo "⚡ 编译中..."
go build -ldflags="-s -w" -o entrag ./cmd/entrag/

# 检查构建是否成功
if [ ! -f "entrag" ]; then
    echo "❌ 构建失败！"
    exit 1
fi

echo "✅ 构建成功！"

# 显示文件大小
echo "📊 文件信息:"
ls -lh entrag

# 更新全局安装
echo "🔄 更新全局安装..."
sudo cp entrag /usr/local/bin/entrag

echo "🎉 优化版本安装完成！"
echo ""
echo "新功能："
echo "  📊 entrag stats    - 查看详细统计"
echo "  🧹 entrag cleanup  - 清理和优化数据库"
echo "  ⚡ entrag optimize - 性能优化和缓存预热"
echo "  🔍 entrag ask      - 带时间统计的智能问答"
echo ""
echo "优化特性："
echo "  ✅ Chunk重叠支持 (提高连续性)"
echo "  ✅ 向量缓存 (避免重复计算)"
echo "  ✅ 并行处理 (3x加速索引构建)"
echo "  ✅ 智能搜索 (文件多样性优化)"
echo "  ✅ 性能监控 (详细时间分析)" 