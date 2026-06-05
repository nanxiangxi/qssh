/**
 * 危险命令测试脚本
 * 用于验证 commandSecurityAnalyzer 是否正确识别各类危险命令
 */

import { commandSecurityAnalyzer } from './commandSecurityAnalyzer.js'

const testCommands = [
  // Critical 级别
  { cmd: ':(){ :|:& };:', expectedLevel: 'critical', desc: 'Fork Bomb' },
  { cmd: 'rm -rf /', expectedLevel: 'critical', desc: '删除根目录' },
  { cmd: 'rm -rf /*', expectedLevel: 'critical', desc: '删除根目录所有文件' },
  { cmd: 'dd if=/dev/zero of=/dev/sda', expectedLevel: 'critical', desc: '清空磁盘' },
  { cmd: 'mkfs.ext4 /dev/sda1', expectedLevel: 'critical', desc: '格式化分区' },
  
  // High 级别
  { cmd: '> /etc/passwd', expectedLevel: 'high', desc: '清空密码文件' },
  { cmd: 'chmod 777 /etc', expectedLevel: 'high', desc: '开放系统目录权限' },
  { cmd: 'iptables -F', expectedLevel: 'high', desc: '清空防火墙规则' },
  { cmd: 'kill -9 -1', expectedLevel: 'high', desc: '杀死所有进程' },
  
  // Medium 级别
  { cmd: 'cat /etc/shadow', expectedLevel: 'medium', desc: '读取影子密码' },
  { cmd: 'sudo su', expectedLevel: 'medium', desc: '切换到 root' },
  
  // Safe 级别
  { cmd: 'ls', expectedLevel: 'safe', desc: '列出文件' },
  { cmd: 'pwd', expectedLevel: 'safe', desc: '显示当前目录' },
  { cmd: 'cat file.txt', expectedLevel: 'safe', desc: '查看文件' },
]

console.log('🧪 开始测试危险命令检测...\n')

let passed = 0
let failed = 0

testCommands.forEach(({ cmd, expectedLevel, desc }) => {
  const result = commandSecurityAnalyzer.analyzeCommand(cmd)
  const isPassed = result.riskLevel === expectedLevel
  
  if (isPassed) {
    console.log(`✅ [${expectedLevel.toUpperCase()}] ${desc}`)
    console.log(`   命令: ${cmd}`)
    passed++
  } else {
    console.log(`❌ [期望: ${expectedLevel}, 实际: ${result.riskLevel}] ${desc}`)
    console.log(`   命令: ${cmd}`)
    console.log(`   详情:`, result)
    failed++
  }
  console.log()
})

console.log('\n' + '='.repeat(60))
console.log(`测试结果: ${passed} 通过, ${failed} 失败, 共 ${testCommands.length} 个`)
console.log('='.repeat(60))

if (failed === 0) {
  console.log('\n🎉 所有测试通过！')
} else {
  console.log(`\n⚠️ 有 ${failed} 个测试失败，请检查`)
  process.exit(1)
}
