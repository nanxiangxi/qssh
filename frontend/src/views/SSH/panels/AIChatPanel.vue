<template>
  <div class="ai-panel">
    <Message ref="msgRef" />

    <div class="ai-main">
      <!-- 头部 -->
      <div class="ai-header">
        <div class="header-left">
          <span class="header-title">AI 助手</span>
          <span v-if="!configured" class="badge warn">未配置</span>
          <span v-else class="badge ok">{{ config.model }}</span>
        </div>
        <div class="header-right">
          <button class="hbtn" @click="clearAll" title="清除对话">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          </button>
          <button class="hbtn" @click="showCfg=true" title="设置">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          </button>
        </div>
      </div>

      <!-- 消息区域 -->
      <div class="ai-msgs" ref="msgsEl">
        <!-- 空状态 -->
        <div v-if="timeline.length===0 && !loading" class="empty">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.3"><path d="M12 2a7 7 0 0 1 7 7c0 2.38-1.19 4.47-3 5.74V17a2 2 0 0 1-2 2H10a2 2 0 0 1-2-2v-2.26C6.19 13.47 5 11.38 5 9a7 7 0 0 1 7-7z"/><path d="M10 21h4"/><path d="M9 9h6"/><path d="M12 9v4"/></svg>
          <p class="empty-t">SSH AI 助手</p>
          <p class="empty-h">分析服务器状态、执行命令、排查问题</p>
          <div class="shortcuts">
            <button @click="quick('查看当前服务器的系统信息')">系统信息</button>
            <button @click="quick('查看磁盘使用情况')">磁盘状态</button>
            <button @click="quick('查看当前运行的进程')">进程列表</button>
            <button @click="quick('检查服务器安全配置')">安全检查</button>
          </div>
        </div>

        <!-- 时间线 -->
        <template v-for="item in timeline" :key="item.id">
          <!-- 用户消息 -->
          <div v-if="item.kind==='user'" class="bubble user">
            <div class="bbody">
              <div class="bcontent" v-html="md(item.content)"></div>
              <div class="btime">{{ time(item.timestamp) }}</div>
            </div>
            <div class="bavatar u">U</div>
          </div>

          <!-- AI 回复 -->
          <div v-else-if="item.kind==='assistant'" class="bubble ai">
            <div class="bavatar">AI</div>
            <div class="bbody">
              <div class="bcontent" v-html="md(item.content)"></div>
              <div class="btime">{{ time(item.timestamp) }}</div>
            </div>
          </div>

          <!-- 工具结果 -->
          <div v-else-if="item.kind==='tool'" class="tool-card">
            <div class="tc-head"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="icon-success"><polyline points="20 6 9 17 4 12"/></svg>{{ item.command }}</div>
            <pre class="tc-body">{{ item.result || '(无输出)' }}</pre>
          </div>

          <!-- 步骤 -->
          <div v-else-if="item.kind==='steps'" class="steps-card">
            <div class="sc-head" @click="toggleSteps(item.id)">
              <span><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg> 执行过程 ({{ item.steps.length }})</span>
              <svg :class="{flip:isStepsOpen(item.id)}" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
            </div>
            <div v-if="isStepsOpen(item.id)" class="sc-list">
              <div v-for="(s,i) in item.steps" :key="i" class="sc-step" :class="{done:s.done}">
                <span v-if="s.done" class="sc-ok"><svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" class="icon-success"><polyline points="20 6 9 17 4 12"/></svg></span>
                <span v-else class="sc-spin"></span>
                {{ s.text }}
              </div>
            </div>
          </div>
        </template>

        <!-- 深度思考推理内容（实时） -->
        <div v-if="reasoning" class="reasoning-card">
          <div class="rc-head">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="icon-purple"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
            <span>深度思考</span>
          </div>
          <div class="rc-body">{{ reasoning }}</div>
        </div>
      </div>

      <!-- 审批栏 -->
      <div v-if="approvals.length" class="approval">
        <div v-for="a in approvals" :key="a.id" class="ap-item">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="icon-warning"><path d="M12 9v4"/><path d="M12 17h.01"/><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/></svg>
          <code class="ap-cmd">{{ a.command }}</code>
          <div class="ap-btns">
            <button class="ap-deny" @click="deny(a.id)">拒绝</button>
            <button class="ap-ok" @click="approve(a.id)">执行</button>
          </div>
        </div>
      </div>

      <!-- 输入区 -->
      <div class="ai-input">
        <div class="in-bar">
          <select class="msel" v-model="targetTerminal" @change="onTerminalChange">
            <option value="">AI 自动</option>
            <option v-for="t in terminalList" :key="t.id" :value="t.id">{{ t.title }}</option>
          </select>
          <div class="bar-spacer"></div>
          <select class="msel" @change="runQuick($event.target.value); $event.target.value=''">
            <option value="" disabled selected>快捷功能</option>
            <optgroup label="运维操作">
              <option value="clean_cache">清理服务器缓存</option>
              <option value="update_patches">检查系统补丁</option>
              <option value="data_backup">数据备份方案</option>
              <option value="log_cleanup">清理过期日志</option>
            </optgroup>
            <optgroup label="安全检查">
              <option value="security_audit">安全审计</option>
              <option value="vuln_scan">漏洞扫描</option>
              <option value="ssh_security">SSH 安全检查</option>
              <option value="firewall_audit">防火墙审计</option>
            </optgroup>
            <optgroup label="性能分析">
              <option value="disk_analysis">磁盘分析</option>
              <option value="process_analysis">进程分析</option>
              <option value="network_analysis">网络连接分析</option>
              <option value="boot_analysis">启动服务分析</option>
            </optgroup>
            <optgroup label="报告生成">
              <option value="operation_report">运维报告</option>
              <option value="terminal_review">操作复盘</option>
            </optgroup>
          </select>
          <select v-if="models.length" class="msel" :value="config.model" @change="switchModel($event.target.value)">
            <option v-for="m in models" :key="m.id" :value="m.id">{{ m.id }}</option>
          </select>
        </div>
        <div class="in-row">
          <textarea v-model="input" @keydown.enter.exact.prevent="send" @input="autoH" placeholder="输入问题... (Enter 发送)" rows="1" ref="inEl"></textarea>
          <button v-if="loading" class="cancel-btn" @click="cancelProcessing" title="取消">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="6" y="6" width="12" height="12" rx="2"/></svg>
          </button>
          <button v-else class="send" @click="send" :disabled="!input.trim()||!configured">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
          </button>
        </div>
        <div class="in-hint">
          <span v-if="!configured" class="hw">请先配置 AI API</span>
          <span v-else class="ht">{{ config.model }}</span>
        </div>
      </div>
    </div>

    <!-- 设置弹窗 -->
    <Teleport to="body">
      <div v-if="showCfg" class="cfg-mask" @click.self="showCfg=false">
        <div class="cfg-box">
          <div class="cfg-head"><h3>AI 设置</h3><button @click="showCfg=false">&times;</button></div>
          <div class="cfg-body">
            <div class="cfg-sec">
              <div class="cfg-sl">连接</div>
              <label>API 端点 <small>填 base URL 或完整 URL 均可</small></label>
              <div class="cfg-row">
                <input v-model="cfgForm.api_endpoint" placeholder="https://api.openai.com/v1" />
                <button class="cfg-fetch" @click="fetchModels" :disabled="!cfgForm.api_endpoint||!cfgForm.api_key">获取模型</button>
              </div>
              <label>API 密钥</label>
              <input v-model="cfgForm.api_key" type="password" placeholder="sk-..." />
            </div>
            <div class="cfg-sec">
              <div class="cfg-sl">模型</div>
              <div v-if="models.length" class="cfg-models">
                <button v-for="m in models" :key="m.id" class="cfg-mc" :class="{on:cfgForm.model===m.id}" @click="cfgForm.model=m.id">{{ m.id }}</button>
              </div>
              <label v-else>模型名称</label>
              <input v-if="!models.length" v-model="cfgForm.model" placeholder="模型名称（如 gpt-4, mimo-v2.5-pro）" />
              <div class="cfg-presets">
                <button v-for="(_,k) in PRESETS" :key="k" @click="applyPreset(k)">{{ k }}</button>
              </div>
            </div>
            <div class="cfg-sec">
              <div class="cfg-sl">参数</div>
              <div class="cfg-2col">
                <div><label>温度</label><input v-model.number="cfgForm.temperature" type="number" step="0.1" min="0" max="2" /></div>
                <div><label>Tokens</label><input v-model.number="cfgForm.max_tokens" type="number" min="100" max="131072" /></div>
              </div>
            </div>
            <div class="cfg-sec">
              <div class="cfg-sl">系统提示词</div>
              <textarea v-model="cfgForm.system_prompt" rows="3"></textarea>
            </div>
          </div>
          <div class="cfg-foot">
            <button class="cfg-cancel" @click="showCfg=false">取消</button>
            <button class="cfg-save" @click="saveCfg" :disabled="saving">{{ saving?'保存中...':'保存' }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, inject } from 'vue'
import { AIService } from '../../../../bindings/changeme/ai/index.js'
import { ChatRequest, AIConfig } from '../../../../bindings/changeme/ai/models.js'
import { Events } from '@wailsio/runtime'
import { marked } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/github-dark-dimmed.css'
import Message from '../../../components/Message.vue'
import { useAIToolLogStore } from '../../../stores/aiToolLog'

// --- marked 配置 ---
const mdRenderer = new marked.Renderer()
mdRenderer.code = (code, lang) => {
  const t = typeof code === 'object' ? code.text : code
  const l = typeof code === 'object' ? code.lang : lang
  const hi = l && hljs.getLanguage(l) ? hljs.highlight(t, {language: l}).value : hljs.highlightAuto(t).value
  const cmds = ['cat','ls','df','free','top','ps','who','last','uname','hostname','uptime','ss ','systemctl','tail','head','grep','find','docker','curl','wget','echo','awk','wc','mount']
  const isCmd = ['bash','sh','shell','zsh'].includes(l) || (t.split('\n').filter(x=>x.trim()&&!x.trim().startsWith('#')).length===1 && cmds.some(p=>t.trim().startsWith(p)))
  const btn = isCmd ? `<button class="ce-btn" data-c="${t.replace(/"/g,'&quot;')}">执行</button>` : ''
  return `<div class="cb"><div class="cb-h">${l||'code'}${btn}</div><pre><code class="hljs">${hi}</code></pre></div>`
}
marked.setOptions({ breaks: true, gfm: true })
const md = (t) => t ? marked.parse(t, {renderer: mdRenderer}) : ''

const PRESETS = {
  'MiniMax': { api_endpoint:'https://api.mc-plfd-host.top/v1', model:'MiniMax-M2.1-highspeed' },
  'MiMo': { api_endpoint:'https://token-plan-cn.xiaomimimo.com/v1', model:'mimo-v2.5-pro' },
  'OpenAI': { api_endpoint:'https://api.openai.com/v1', model:'gpt-4' },
  '火山豆包': { api_endpoint:'https://ark.cn-beijing.volces.com/api/v3', model:'doubao-seed-1-6-251015' },
  'Ollama': { api_endpoint:'http://localhost:11434/v1', api_key:'ollama', model:'llama3.2' },
}

// --- 状态 ---
const connId = inject('connId')
const toolLog = useAIToolLogStore()
const msgRef = ref(null)
const msgsEl = ref(null)
const inEl = ref(null)

const msgs = ref([])        // {id, role, content, timestamp}
const input = ref('')
const loading = ref(false)
const configured = ref(false)
const terminalList = ref([])
const targetTerminal = ref('') // '' = AI 自动
const showCfg = ref(false)
const saving = ref(false)
const models = ref([])

const config = ref({ api_endpoint:'', api_key:'', model:'', timeout:120, temperature:1.0, max_tokens:4096, system_prompt:'' })
const cfgForm = ref({ ...config.value })

// 当前轮次状态
const steps = ref([])
const reasoning = ref('') // 深度思考推理内容
const approvals = ref([])
const toolOuts = ref([])
const triggerId = ref(null)

// 已归档轮次
const archived = ref([]) // [{tid, steps, tools}]
const stepsOpen = ref({}) // { [stepId]: boolean } — 展开状态独立存储

const toggleSteps = (id) => {
  stepsOpen.value[id] = !stepsOpen.value[id]
}

const isStepsOpen = (id) => {
  return stepsOpen.value[id] !== false // 默认展开
}

// 时间线
const timeline = computed(() => {
  const items = []
  for (const m of msgs.value) {
    items.push({ ...m, kind: m.role, id: m.id })
    const a = archived.value.find(x => x.tid === m.id)
    if (a) {
      if (a.steps.length) {
        const sid = `s_${m.id}`
        items.push({ kind:'steps', id:sid, steps:a.steps })
      }
      a.tools.forEach(t => items.push({ kind:'tool', id:t.id, command:t.command, result:t.result }))
    }
  }
  if (steps.value.length && triggerId.value && !archived.value.find(x => x.tid === triggerId.value)) {
    items.push({ kind:'steps', id:'s_cur', steps:steps.value })
  }
  if (toolOuts.value.length && triggerId.value && !archived.value.find(x => x.tid === triggerId.value)) {
    toolOuts.value.forEach(t => items.push({ kind:'tool', id:t.id, command:t.command, result:t.result }))
  }
  return items
})

// --- Markdown 工具 ---
const time = (ts) => ts ? new Date(ts).toLocaleTimeString('zh-CN',{hour:'2-digit',minute:'2-digit'}) : ''
const autoH = (e) => { e.target.style.height='auto'; e.target.style.height=Math.min(e.target.scrollHeight,120)+'px' }
const scroll = (d=0) => { nextTick(()=>{ if(msgsEl.value) msgsEl.value.scrollTop=msgsEl.value.scrollHeight; if(d>0) setTimeout(()=>{ if(msgsEl.value) msgsEl.value.scrollTop=msgsEl.value.scrollHeight },d) }) }

// --- 发送 ---
const send = async () => {
  const t = input.value.trim()
  if (!t || !connId || !configured.value) return
  const tid = `u_${Date.now()}`
  msgs.value.push({ id:tid, role:'user', content:t, timestamp:new Date().toISOString() })
  input.value = ''
  if(inEl.value) inEl.value.style.height='auto'
  await scroll()
  try {
    const ctx = toolLog.getSessionSummary(connId)
    let msg = ctx ? t + '\n\n---\n' + ctx : t
    const resp = await AIService.SendMessage(new ChatRequest({ connId, message: msg }))
    if (resp && resp.message === '已排队等待处理') {
      msgRef.value?.info('已排队等待处理')
    } else {
      // 新消息开始处理，更新 triggerId
      triggerId.value = tid
      steps.value.splice(0); approvals.value.splice(0); toolOuts.value.splice(0)
      loading.value = true
    }
  } catch(e) {
    msgRef.value?.error('发送失败')
  }
}

// 取消当前 AI 处理
const cancelProcessing = async () => {
  try {
    await AIService.CancelProcessing(connId)
    loading.value = false
    approvals.value.splice(0)
    steps.value.splice(0)
    msgRef.value?.info('已取消')
  } catch(e) {
    console.error('取消失败:', e)
  }
}

const quick = (t) => { input.value=t; send() }

// 快捷功能
const QUICK_ACTIONS = {
  // 运维操作
  clean_cache: '请帮我清理服务器缓存和临时文件：清理系统缓存（/tmp、/var/tmp）、清理包管理器缓存（yum clean all 或 apt clean）、清理旧日志文件、清理 Docker 无用镜像和容器（如已安装）。列出每步操作和释放的空间。',
  update_patches: '请检查当前服务器是否有待安装的系统安全补丁和更新：检查可用更新列表、区分安全更新和普通更新、评估更新风险、给出更新建议和操作步骤。',
  data_backup: '请帮我制定数据备份方案：检查当前磁盘使用情况、识别关键数据目录（/etc、/home、/var、数据库目录等）、检查是否有现有备份机制、给出备份策略建议（包括增量备份、定时任务配置）。',
  log_cleanup: '请帮我清理过期日志文件：检查 /var/log 目录大小、找出最大的日志文件、检查日志轮转配置、清理超过30天的旧日志、给出日志管理建议。',
  // 安全检查
  security_audit: '请对当前服务器进行全面安全审计：检查最近登录记录、开放的端口、可疑进程、SSH配置安全性、防火墙状态、密码策略、文件权限异常，并给出安全建议。',
  vuln_scan: '请检查当前服务器已安装的关键软件版本（ssh、openssl、nginx、kernel等），根据已知的常见漏洞版本进行对比分析，列出可能存在风险的软件及当前版本，给出修复建议和安全更新命令。',
  ssh_security: '请检查 SSH 服务的安全配置：检查 /etc/ssh/sshd_config 中的危险配置（root登录、密码认证、端口等）、检查 authorized_keys 文件、检查最近的SSH登录失败记录、给出加固建议。',
  firewall_audit: '请检查防火墙配置：查看当前防火墙规则（iptables/firewalld/ufw）、检查开放的端口是否合理、检查是否有不必要的入站规则、给出防火墙加固建议。',
  // 性能分析
  disk_analysis: '请分析当前服务器的磁盘使用情况：查看各分区使用率、找出占用空间最大的目录和文件（du -h --max-depth）、检查 inode 使用情况、给出清理建议。',
  process_analysis: '请分析当前服务器的进程情况：查看 CPU 和内存占用最高的进程、检查是否有异常进程、分析进程资源趋势、给出优化建议。',
  network_analysis: '请分析当前服务器的网络连接状态：查看所有网络连接、检查监听端口、分析连接数和状态分布、检查是否有异常外部连接、给出网络安全建议。',
  boot_analysis: '请分析当前服务器的启动服务：列出所有开机自启服务、检查各服务状态、分析启动耗时、识别不必要的自启服务、给出优化建议。',
  // 报告生成
  operation_report: '请生成一份完整的服务器运维报告：包括系统信息、资源使用情况、磁盘状态、运行的服务、安全状态、最近的登录记录，给出综合评估和改进建议。',
  terminal_review: '请分析当前服务器最近在终端执行的操作：查看 bash_history、最近修改的文件、最近的 sudo 操作日志，总结执行了哪些命令、做了什么变更，并给出操作建议。'
}

const runQuick = (action) => {
  const text = QUICK_ACTIONS[action]
  if (text) {
    input.value = text
    send()
  }
}

// --- 事件 ---
const onMsg = (ev) => {
  const d = ev?.data; if(!d||d.connId!==connId||!d.message) return
  if(msgs.value.find(m=>m.id===d.message.id)) return
  msgs.value.push(d.message)
  loading.value = false
  if(steps.value.length) { steps.value[steps.value.length-1].done=true; archived.value.push({tid:triggerId.value, steps:[...steps.value], tools:[...toolOuts.value], open:true}); steps.value.splice(0) }
  toolOuts.value.splice(0)
  reasoning.value = ''
  triggerId.value = null
  scroll(200)
}

const onStatus = (ev) => {
  const d = ev?.data; if(!d||d.connId!==connId) return
  loading.value = true
  if(d.status==='step' && d.text) {
    if(steps.value.length) steps.value[steps.value.length-1].done=true
    steps.value.push({text:d.text, done:false})
  }
  scroll()
}

// 流式内容：推理过程和正文
const onStreamChunk = (ev) => {
  const d = ev?.data; if(!d||d.connId!==connId) return
  if (d.type === 'reasoning' && d.content) {
    reasoning.value += d.content
    scroll()
  }
  // content 类型的 chunk 不需要特殊处理，最终消息到达时一次性显示
}

const onApproval = (ev) => {
  const d = ev?.data; if(!d||d.connId!==connId) return
  toolLog.logToolCall(connId, d.callId, d.tool, d.command)
  approvals.value.push({id:d.callId, tool:d.tool, command:d.command})
  loading.value = false
  scroll()
}

const onResult = (ev) => {
  const d = ev?.data; if(!d||d.connId!==connId) return
  toolLog.logToolResult(connId, d.callId, d.result)
  approvals.value = approvals.value.filter(a=>a.id!==d.callId)
  toolOuts.value.push({id:d.callId, command:d.command, result:d.result})
  scroll()
}

const approve = async (id) => {
  toolLog.logToolApprove(connId, id)
  await AIService.ApproveTool(id)
  const a = approvals.value.find(x=>x.id===id); if(a) a.done=true
}

const deny = async (id) => {
  toolLog.logToolDeny(connId, id)
  await AIService.DenyTool(id)
  approvals.value = approvals.value.filter(a=>a.id!==id)
}

// --- 模型 ---
const fetchModels = async () => {
  try {
    // 使用表单中当前填写的 endpoint 和 key，而非已保存的配置
    const r = await AIService.FetchModelsWithParams(cfgForm.value.api_endpoint, cfgForm.value.api_key)
    if (r?.length) {
      models.value = r
      msgRef.value?.success(`获取到 ${r.length} 个模型`)
    } else {
      msgRef.value?.error('未获取到模型，请检查 API 端点和密钥')
    }
  } catch (e) {
    msgRef.value?.error('获取模型失败: ' + (e.message || e))
  }
}

const switchModel = async (m) => {
  config.value.model = m
  cfgForm.value.model = m
  try { await AIService.SaveConfig(new AIConfig({...config.value})); msgRef.value?.success('已切换 '+m) } catch(e) { msgRef.value?.error('切换失败') }
}

// --- 设置 ---
const applyPreset = (k) => {
  const p = PRESETS[k]
  if (p) {
    cfgForm.value.api_endpoint = p.api_endpoint || cfgForm.value.api_endpoint
    cfgForm.value.model = p.model || cfgForm.value.model
    if (p.api_key) cfgForm.value.api_key = p.api_key
  }
}

const loadCfg = async () => {
  try {
    const c = await AIService.GetConfig()
    if(c) {
      config.value = { api_endpoint:c.api_endpoint||'', api_key:c.api_key||'', model:c.model||'', timeout:c.timeout||120, temperature:c.temperature||1.0, max_tokens:c.max_tokens||4096, system_prompt:c.system_prompt||'' }
      cfgForm.value = { ...config.value }
      configured.value = await AIService.IsConfigured()
    }
  } catch(e) {}
}

const saveCfg = async () => {
  if(!cfgForm.value.api_endpoint||!cfgForm.value.api_key||!cfgForm.value.model) { msgRef.value?.error('请填写完整配置'); return }
  saving.value = true
  try {
    await AIService.SaveConfig(new AIConfig({...cfgForm.value}))
    config.value = { ...cfgForm.value }
    configured.value = true; showCfg.value = false
    msgRef.value?.success('配置已保存')
  } catch(e) { msgRef.value?.error('保存失败') }
  finally { saving.value = false }
}

const loadHistory = async () => {
  if(!connId) return
  try { const h = await AIService.GetChatHistory(connId); if(h?.length) { msgs.value=h; scroll() } } catch(e) {}
}

const clearAll = async () => {
  if(!connId||!msgs.value.length) return
  try {
    await AIService.ClearChatHistory(connId)
    msgs.value.splice(0); steps.value.splice(0); approvals.value.splice(0); toolOuts.value.splice(0); archived.value.splice(0); triggerId.value=null
    toolLog.clearLogs(connId)
    msgRef.value?.success('已清除')
  } catch(e) { msgRef.value?.error('清除失败') }
}

// --- 生命周期 ---
// 终端选择变更
const onTerminalChange = () => {
  import('../../../utils/aiToolExecutor.js').then(mod => {
    // 找到选中终端的 sessionId
    const selected = terminalList.value.find(t => t.id === targetTerminal.value)
    mod.setTargetTerminal(selected?.sessionId || null)
  })
}

onMounted(async () => {
  await loadCfg(); await loadHistory(); fetchModels()
  Events.On('ai:message', onMsg)
  Events.On('ai:status', onStatus)
  Events.On('ai:stream-chunk', onStreamChunk)
  Events.On('ai:tool-approval', onApproval)
  Events.On('ai:tool-result', onResult)
  // 监听终端列表变化
  Events.On('dockview:terminals-changed', (e) => {
    if (e?.data?.terminals) terminalList.value = e.data.terminals
  })
  if(inEl.value) inEl.value.focus()
  // 代码块执行按钮事件委托
  msgsEl.value?.addEventListener('click', e => {
    const btn = e.target.closest('.ce-btn')
    if(btn) { input.value=btn.dataset.c; send() }
  })
})

onUnmounted(() => {
  Events.Off('ai:message', onMsg)
  Events.Off('ai:status', onStatus)
  Events.Off('ai:stream-chunk', onStreamChunk)
  Events.Off('ai:tool-approval', onApproval)
  Events.Off('dockview:terminals-changed')
  Events.Off('ai:tool-result', onResult)
})
</script>

<style>
/* 全局：代码块执行按钮（Teleport 外需要非 scoped） */
.ce-btn { padding:2px 8px; background:var(--success-bg, rgba(72, 187, 120, 0.15)); border:1px solid var(--border-success, rgba(72, 187, 120, 0.3)); border-radius:3px; color:var(--success-light); font-size:10px; cursor:pointer; transition:all .15s }
.ce-btn:hover { background:var(--success-bg) }
</style>

<style scoped>
.ai-panel { width:100%; height:100%; display:flex; background:var(--bg-panel); overflow:hidden; font-size:13px }
.ai-main { flex:1; display:flex; flex-direction:column; min-height:0; overflow:hidden }

/* 头部 */
.ai-header { display:flex; justify-content:space-between; align-items:center; padding:10px 14px; background:var(--bg-tooltip, rgba(28, 28, 28, 0.98)); border-bottom:1px solid var(--border-subtle); flex-shrink:0 }
.header-left { display:flex; align-items:center; gap:8px }
.header-title { color:var(--text-primary); font-weight:600; font-size:13px }
.badge { font-size:10px; padding:2px 6px; border-radius:3px; font-family:monospace }
.badge.warn { color:var(--accent-danger); background:var(--danger-bg, rgba(252, 129, 129, 0.1)) }
.badge.ok { color:var(--success-light); background:var(--success-bg, rgba(72, 187, 120, 0.1)) }
.header-right { display:flex; gap:2px }
.hbtn { background:none; border:none; color:var(--text-muted); cursor:pointer; padding:6px; border-radius:4px; display:flex; transition:all .15s }
.hbtn:hover { background:var(--border-subtle); color:var(--text-primary) }

/* 消息区 */
.ai-msgs { flex:1; overflow-y:auto; padding:16px; display:flex; flex-direction:column; gap:14px }
.ai-msgs::-webkit-scrollbar { width:5px }
.ai-msgs::-webkit-scrollbar-thumb { background:var(--surface-hover); border-radius:3px }

/* 空状态 */
.empty { flex:1; display:flex; flex-direction:column; align-items:center; justify-content:center; text-align:center }
.empty-t { color:var(--text-secondary); font-size:15px; font-weight:600; margin:12px 0 4px }
.empty-h { color:var(--text-disabled); font-size:12px; margin:0 0 20px }
.shortcuts { display:flex; flex-wrap:wrap; gap:8px; justify-content:center }
.shortcuts button { padding:7px 14px; background:var(--surface-1); border:1px solid var(--border-default); border-radius:8px; color:var(--text-secondary); font-size:12px; cursor:pointer; transition:all .15s }
.shortcuts button:hover { background:var(--primary-bg); border-color:var(--border-accent); color:var(--primary-light) }

/* 消息气泡 */
.bubble { display:flex; gap:10px; animation:msgIn .25s ease-out }
@keyframes msgIn { from{opacity:0;transform:translateY(8px)} to{opacity:1;transform:translateY(0)} }
.bubble.user { justify-content:flex-end }
.bavatar { width:30px; height:30px; border-radius:8px; display:flex; align-items:center; justify-content:center; flex-shrink:0; font-size:10px; font-weight:700; background:var(--primary-bg, rgba(66, 153, 225, 0.12)); color:var(--primary-light) }
.bavatar.u { background:var(--success-bg, rgba(72, 187, 120, 0.12)); color:var(--success-light) }
.bbody { max-width:80%; min-width:0; display:flex; flex-direction:column }
.bubble.user .bbody { align-items:flex-end }
.bcontent { padding:10px 14px; border-radius:10px; background:var(--bg-panel); color:var(--text-primary); line-height:1.7; word-break:break-word; border:1px solid var(--surface-1) }
.bubble.user .bcontent { background:var(--primary-bg); border-color:var(--primary-bg) }
.btime { font-size:10px; color:var(--text-disabled); margin-top:3px }

/* Markdown 内容 */
.bcontent :deep(h2),.bcontent :deep(h3),.bcontent :deep(h4) { color:var(--text-primary); font-weight:600 }
.bcontent :deep(h2) { font-size:15px; margin:14px 0 6px }
.bcontent :deep(h3) { font-size:14px; margin:12px 0 4px }
.bcontent :deep(h4) { font-size:13px; margin:10px 0 4px }
.bcontent :deep(ul),.bcontent :deep(ol) { margin:6px 0; padding-left:20px }
.bcontent :deep(li) { margin:3px 0; line-height:1.6 }
.bcontent :deep(strong) { color:var(--text-primary) }
.bcontent :deep(blockquote) { border-left:3px solid var(--border-accent); padding:4px 12px; margin:8px 0; color:var(--text-secondary); background:var(--primary-bg, rgba(66, 153, 225, 0.06)); border-radius:0 4px 4px 0 }
.bcontent :deep(a) { color:var(--primary-light); text-decoration:none }
.bcontent :deep(hr) { border:none; border-top:1px solid var(--border-subtle); margin:12px 0 }
.bcontent :deep(table) { width:100%; border-collapse:collapse; font-size:12px; margin:8px 0; background:var(--surface-1); border-radius:6px; overflow:hidden }
.bcontent :deep(th) { padding:6px 10px; text-align:left; color:var(--text-primary); font-weight:600; background:var(--surface-1); border-bottom:1px solid var(--border-default) }
.bcontent :deep(td) { padding:5px 10px; border-bottom:1px solid var(--surface-1); color:var(--text-secondary) }
.bcontent :deep(tr:last-child td) { border-bottom:none }
.bcontent :deep(.cb) { margin:8px 0; border-radius:6px; overflow:hidden; background:var(--bg-panel-solid); border:1px solid var(--border-subtle) }
.bcontent :deep(.cb-h) { display:flex; align-items:center; justify-content:space-between; font-size:10px; color:var(--text-muted); padding:4px 10px; background:var(--surface-1); border-bottom:1px solid var(--border-subtle); text-transform:uppercase; font-family:monospace }
.bcontent :deep(pre) { margin:0; padding:10px; overflow-x:auto; background:transparent }
.bcontent :deep(pre code) { font-family:'Cascadia Code','Fira Code',Consolas,monospace; font-size:12px; line-height:1.5 }
.bcontent :deep(code) { font-family:'Cascadia Code','Fira Code',Consolas,monospace; font-size:12px }
.bcontent :deep(.inline-code) { background:var(--shadow-sm); padding:1px 5px; border-radius:3px; font-size:.9em }

/* 工具结果 */
.tool-card { border-radius:8px; background:var(--bg-panel); border:1px solid var(--surface-1); animation:msgIn .25s ease-out; overflow:visible }
.tc-head { display:flex; align-items:center; gap:6px; padding:6px 12px; background:var(--success-bg, rgba(72, 187, 120, 0.06)); border-bottom:1px solid var(--surface-1); color:var(--text-secondary); font-size:11px; font-family:monospace }
.tc-body { margin:0; padding:8px 12px; color:var(--text-secondary); font-size:12px; font-family:'Cascadia Code','Fira Code',monospace; line-height:1.5; min-height:24px; max-height:180px; overflow-y:auto; white-space:pre-wrap; word-break:break-all }
.tc-body::-webkit-scrollbar { width:4px }
.tc-body::-webkit-scrollbar-thumb { background:var(--surface-hover); border-radius:2px }

/* 步骤 */
.steps-card { border-radius:8px; background:var(--bg-panel); border:1px solid var(--surface-1); border-left:3px solid var(--border-accent); animation:msgIn .25s ease-out }
.sc-head { display:flex; align-items:center; justify-content:space-between; padding:8px 14px; cursor:pointer; transition:background .15s }
.sc-head:hover { background:var(--surface-1, rgba(255, 255, 255, 0.02)) }
.sc-head span { display:flex; align-items:center; gap:6px; color:var(--text-secondary); font-size:11px; font-weight:600 }
.sc-head svg { color:var(--text-muted); transition:transform .2s }
.sc-head svg.flip { transform:rotate(180deg) }
.sc-list { padding:4px 14px 8px }
.sc-step { display:flex; align-items:center; gap:8px; padding:4px 0; color:var(--text-secondary); font-size:12px; line-height:1.4 }
.sc-step.done { color:var(--text-secondary) }
.sc-ok { display:flex; align-items:center }
.sc-spin { width:10px; height:10px; border:1.5px solid var(--border-accent); border-top-color:var(--primary-light); border-radius:50%; animation:spin .8s linear infinite }
@keyframes spin { to{transform:rotate(360deg)} }

/* 深度思考推理内容 */
.reasoning-card { border-radius:8px; background:var(--bg-panel); border:1px solid var(--accent-purple); border-left:3px solid var(--accent-purple); animation:msgIn .25s ease-out }
.rc-head { display:flex; align-items:center; gap:6px; padding:6px 14px; color:var(--accent-purple); font-size:11px; font-weight:600 }
.rc-body { padding:4px 14px 10px; color:var(--text-secondary); font-size:12px; line-height:1.6; white-space:pre-wrap; max-height:200px; overflow-y:auto }
.rc-body::-webkit-scrollbar { width:4px }
.rc-body::-webkit-scrollbar-thumb { background:var(--surface-hover); border-radius:2px }

/* 审批栏 */
.approval { padding:8px 14px; background:var(--warning-bg); border-top:1px solid var(--warning-bg); flex-shrink:0 }
.ap-item { display:flex; align-items:center; gap:8px; padding:6px 0 }
.ap-item+.ap-item { border-top:1px solid var(--surface-1) }
.ap-cmd { flex:1; color:var(--text-primary); font-family:'Cascadia Code','Fira Code',monospace; font-size:12px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap }
.ap-btns { display:flex; gap:6px; flex-shrink:0 }
.ap-deny { padding:4px 12px; background:var(--danger-bg, rgba(245, 101, 101, 0.1)); border:1px solid var(--danger-bg); border-radius:5px; color:var(--accent-danger); font-size:12px; cursor:pointer; transition:all .15s }
.ap-deny:hover { background:var(--danger-bg) }
.ap-ok { padding:4px 12px; background:var(--success-bg, rgba(72, 187, 120, 0.15)); border:1px solid var(--border-success, rgba(72, 187, 120, 0.3)); border-radius:5px; color:var(--success-light); font-size:12px; cursor:pointer; transition:all .15s }
.ap-ok:hover { background:var(--success-bg, rgba(72, 187, 120, 0.25)) }

/* 输入区 */
.ai-input { padding:12px 14px; background:var(--bg-tooltip, rgba(28, 28, 28, 0.98)); border-top:1px solid var(--border-subtle); flex-shrink:0 }
.in-bar { display:flex; align-items:center; gap:6px; margin-bottom:8px }
.bar-spacer { flex:1 }
.chip { display:flex; align-items:center; gap:4px; padding:3px 10px; background:var(--surface-1); border:1px solid var(--border-default); border-radius:6px; color:var(--text-muted); font-size:11px; cursor:pointer; transition:all .15s }
.chip:hover { background:var(--border-subtle); color:var(--text-secondary) }
.chip.on { background:var(--primary-bg); border-color:var(--border-accent); color:var(--primary-light) }
.msel { padding:3px 8px; background:var(--bg-panel); border:1px solid var(--border-default); border-radius:6px; color:var(--text-secondary); font-size:11px; font-family:monospace; outline:none; cursor:pointer; max-width:180px; margin-left:auto }
.msel:focus { border-color:var(--border-accent) }
.msel option { background:var(--bg-panel-solid); color:var(--text-primary); padding:4px }
.msel optgroup { background:var(--bg-panel-solid); color:var(--text-muted); font-weight:600 }
.in-row { display:flex; gap:8px; align-items:flex-end }
.in-row textarea { flex:1; background:var(--bg-panel); border:1px solid var(--border-default); border-radius:8px; padding:10px 12px; color:var(--text-primary); font-size:13px; font-family:inherit; resize:none; outline:none; max-height:120px; transition:border-color .15s; line-height:1.5 }
.in-row textarea:focus { border-color:var(--border-accent) }
.in-row textarea::placeholder { color:var(--text-disabled) }
.send { width:38px; height:38px; background:var(--primary-bg); border:none; border-radius:8px; color:var(--primary-light); cursor:pointer; display:flex; align-items:center; justify-content:center; transition:all .15s; flex-shrink:0 }
.send:hover:not(:disabled) { background:var(--primary-bg-hover) }
.send:disabled { opacity:.3; cursor:not-allowed }
.cancel-btn { width:38px; height:38px; background:var(--danger-bg, rgba(245, 101, 101, 0.15)); border:1px solid var(--border-danger, rgba(245, 101, 101, 0.3)); border-radius:8px; color:var(--accent-danger); cursor:pointer; display:flex; align-items:center; justify-content:center; transition:all .15s; flex-shrink:0 }
.cancel-btn:hover { background:var(--danger-bg) }
.in-hint { margin-top:5px; padding:0 2px }
.hw { color:var(--accent-danger); font-size:11px }
.ht { color:var(--text-disabled); font-size:11px }

/* 设置弹窗 */
.cfg-mask { position:fixed; inset:0; background:var(--bg-overlay); display:flex; align-items:center; justify-content:center; z-index:10000; backdrop-filter:blur(4px) }
.cfg-box { background:var(--bg-panel-solid); border:1px solid var(--border-default); border-radius:12px; width:480px; max-height:80vh; display:flex; flex-direction:column; box-shadow:0 24px 64px var(--shadow-lg) }
.cfg-head { display:flex; align-items:center; justify-content:space-between; padding:16px 20px; border-bottom:1px solid var(--border-subtle) }
.cfg-head h3 { margin:0; color:var(--text-primary); font-size:15px; font-weight:600 }
.cfg-head button { background:none; border:none; color:var(--text-muted); font-size:22px; cursor:pointer; padding:0; line-height:1 }
.cfg-head button:hover { color:var(--text-primary) }
.cfg-body { flex:1; overflow-y:auto; padding:16px 20px }
.cfg-sec { margin-bottom:18px }
.cfg-sec:last-child { margin-bottom:0 }
.cfg-sl { color:var(--text-primary); font-size:12px; font-weight:600; margin-bottom:10px; padding-bottom:5px; border-bottom:1px solid var(--border-subtle); text-transform:uppercase; letter-spacing:.5px }
.cfg-body label { display:block; color:var(--text-secondary); font-size:12px; font-weight:500; margin:10px 0 5px }
.cfg-body label small { color:var(--text-disabled); font-weight:normal; margin-left:4px }
.cfg-body input,.cfg-body textarea { width:100%; background:var(--bg-panel); border:1px solid var(--border-default); border-radius:6px; padding:8px 10px; color:var(--text-primary); font-size:13px; font-family:inherit; outline:none; box-sizing:border-box; transition:border-color .15s }
.cfg-body input:focus,.cfg-body textarea:focus { border-color:var(--border-accent) }
.cfg-body textarea { resize:vertical; min-height:60px; font-size:12px; line-height:1.6 }
.cfg-row { display:flex; gap:8px }
.cfg-row input { flex:1 }
.cfg-fetch { display:flex; align-items:center; gap:4px; padding:0 12px; white-space:nowrap; background:var(--primary-bg); border:1px solid var(--border-accent); border-radius:6px; color:var(--primary-light); font-size:12px; cursor:pointer; transition:all .15s }
.cfg-fetch:hover:not(:disabled) { background:var(--primary-bg-hover) }
.cfg-fetch:disabled { opacity:.4; cursor:not-allowed }
.cfg-models { display:flex; flex-wrap:wrap; gap:6px; margin-bottom:10px }
.cfg-mc { padding:6px 10px; background:var(--surface-1); border:1px solid var(--border-default); border-radius:5px; color:var(--text-secondary); font-size:11px; font-family:monospace; cursor:pointer; transition:all .15s }
.cfg-mc:hover { background:var(--primary-bg, rgba(66, 153, 225, 0.08)); border-color:var(--primary-bg) }
.cfg-mc.on { background:var(--primary-bg); border-color:var(--border-accent); color:var(--primary-light) }
.cfg-presets { display:flex; flex-wrap:wrap; gap:6px; margin-top:10px; padding-top:10px; border-top:1px solid var(--surface-1) }
.cfg-presets button { padding:4px 10px; background:var(--surface-1); border:1px solid var(--border-default); border-radius:4px; color:var(--text-secondary); font-size:11px; cursor:pointer; transition:all .15s }
.cfg-presets button:hover { background:var(--primary-bg); border-color:var(--border-accent); color:var(--primary-light) }
.cfg-2col { display:grid; grid-template-columns:1fr 1fr; gap:10px }
.cfg-foot { display:flex; justify-content:flex-end; gap:8px; padding:14px 20px; border-top:1px solid var(--border-subtle) }
.cfg-cancel { padding:7px 16px; background:var(--border-subtle); border:none; border-radius:6px; color:var(--text-secondary); font-size:12px; cursor:pointer; transition:all .15s }
.cfg-cancel:hover { background:var(--surface-hover) }
.cfg-save { padding:7px 16px; background:var(--primary-bg-hover); border:1px solid var(--border-accent); border-radius:6px; color:var(--primary-light); font-size:12px; cursor:pointer; transition:all .15s }
.cfg-save:hover:not(:disabled) { background:var(--border-accent) }
.cfg-save:disabled { opacity:.4; cursor:not-allowed }

/* 图标颜色 */
.icon-warning { color: var(--warning-light, #ed8936); }
.icon-success { color: var(--accent-success, #48bb78); }
.icon-purple { color: var(--accent-purple, #9f7aea); }
</style>
