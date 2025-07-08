#!/bin/bash

echo "🧪 Entrag 性能测试"
echo "=================="

# 测试问题列表
declare -a test_questions=(
    "What is Ent ORM?"
    "How to create database schema?"
    "Entity relationships"
    "PDM是什么？"
    "产品数据管理的定义"
    "PLM和PDM的区别"
)

echo "📊 开始性能测试..."
echo ""

total_start=$(date +%s.%N)

for i in "${!test_questions[@]}"; do
    question="${test_questions[$i]}"
    echo "🔍 测试 $((i+1))/${#test_questions[@]}: $question"
    
    # 运行查询并记录时间
    start_time=$(date +%s.%N)
    
    # 使用timeout防止卡死
    timeout 120s entrag ask "$question" > /dev/null 2>&1
    exit_code=$?
    
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    
    if [ $exit_code -eq 0 ]; then
        echo "   ✅ 完成时间: ${duration}s"
    elif [ $exit_code -eq 124 ]; then
        echo "   ⏰ 超时 (>120s)"
    else
        echo "   ❌ 执行错误"
    fi
    
    echo ""
done

total_end=$(date +%s.%N)
total_duration=$(echo "$total_end - $total_start" | bc)

echo "📈 性能测试完成"
echo "   总耗时: ${total_duration}s"
echo "   平均每个查询: $(echo "$total_duration / ${#test_questions[@]}" | bc -l | cut -c1-5)s"

echo ""
echo "💾 缓存和统计信息:"
entrag stats

echo ""
echo "💡 如果性能仍然较慢，建议："
echo "   1. 运行 'entrag optimize' 预热缓存"
echo "   2. 检查 ollama ps 确认模型已加载"
echo "   3. 考虑使用更小的模型如 gemma2:2b"
echo "   4. 调整 max_similar_chunks 到 2-3" 