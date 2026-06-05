/**
 * SSH 事件管理工具
 * 提供简洁的事件监听接口，避免全局状态污染
 */

import { Events } from '@wailsio/runtime'

/**
 * 监听指定分组的更新事件
 * @param {string} groupID - 分组ID
 * @param {Function} callback - 回调函数，接收 connections 数组
 * @returns {Function} 取消监听的函数
 */
export function listenToGroupUpdates(groupID, callback) {
  const handler = (event) => {
    if (!event || !event.data) return
    
    const data = event.data
    if (data.groupID !== groupID) return
    
    const connections = data.connections || []
    callback(connections)
  }
  
  Events.On('ssh:group-updated', handler)
  
  // 返回取消监听的函数
  return () => {
    Events.Off('ssh:group-updated', handler)
  }
}

/**
 * 监听终端输出事件
 * @param {string} connID - 连接ID
 * @param {Function} callback - 回调函数，接收输出数据
 * @returns {Function} 取消监听的函数
 */
export function listenToTerminalOutput(connID, callback) {
  const handler = (event) => {
    if (!event || !event.data) return
    
    const eventData = event.data
    if (eventData.connID !== connID) return
    if (eventData.data) {
      callback(eventData.data)
    }
  }
  
  Events.On('ssh:terminal-output', handler)
  
  return () => {
    Events.Off('ssh:terminal-output', handler)
  }
}
