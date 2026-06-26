// WebSocket 服务：管理实时连接、自动重连、消息分发、弹窗去重
import { ElNotification } from 'element-plus'
import router from '@/router'

type MessageHandler = (data: any) => void

interface WSMessage {
  type: string
  target_type: string
  target_id?: number
  room?: string
  data: any
  timestamp: string
}

interface NotificationData {
  type: string
  title: string
  content: string
  business_type: string
  business_id: number
  id?: string
  data?: {
    application_id?: number
    approved?: boolean
    applicant?: string
    resource?: string
  }
}

class WebSocketService {
  private ws: WebSocket | null = null
  private url: string = ''
  private reconnectAttempts: number = 0
  private maxReconnectAttempts: number = 10
  private reconnectDelay: number = 3000
  private heartbeatTimer: number | null = null
  private handlers: Map<string, MessageHandler[]> = new Map()
  private connected: boolean = false
  private intentionalClose: boolean = false
  private onUnreadCountChange?: (count: number) => void
  private onPendingCountChange?: (count: number) => void
  // 消息去重：记录最近5分钟内已弹窗的消息ID，防止MQ重投递导致重复弹窗
  private shownMessages: Map<string, number> = new Map()
  private readonly dedupWindowMs = 5 * 60 * 1000

  connect(token: string) {
    this.intentionalClose = false
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    this.url = `${protocol}//${window.location.host}/ws?token=${token}`
    this.doConnect()
  }

  private doConnect() {
    if (this.ws) {
      this.cleanup()
    }

    try {
      this.ws = new WebSocket(this.url)
    } catch (e) {
      console.error('[WS] WebSocket creation failed', e)
      this.scheduleReconnect()
      return
    }

    this.ws.onopen = () => {
      console.log('[WS] Connected')
      this.connected = true
      this.reconnectAttempts = 0
      this.startHeartbeat()
      this.fetchCounts()
    }

    this.ws.onmessage = (event) => {
      if (event.data === 'pong') return
      try {
        const msg: WSMessage = JSON.parse(event.data)
        this.handleMessage(msg)
      } catch (e) {
        console.warn('[WS] Failed to parse message', e)
      }
    }

    this.ws.onerror = (error) => {
      console.error('[WS] Error', error)
    }

    this.ws.onclose = (event) => {
      console.log('[WS] Closed', event.code, event.reason)
      this.connected = false
      this.stopHeartbeat()
      if (!this.intentionalClose) {
        this.scheduleReconnect()
      }
    }
  }

  private isDuplicate(msgId: string): boolean {
    if (!msgId) return false
    this.cleanupExpiredDedups()
    if (this.shownMessages.has(msgId)) {
      console.log('[WS] Duplicate message suppressed:', msgId)
      return true
    }
    this.shownMessages.set(msgId, Date.now())
    return false
  }

  private cleanupExpiredDedups() {
    const now = Date.now()
    for (const [id, ts] of this.shownMessages) {
      if (now - ts > this.dedupWindowMs) {
        this.shownMessages.delete(id)
      }
    }
  }

  private handleMessage(msg: WSMessage) {
    // 心跳响应
    if (msg.type === 'pong') return

    // 统一处理通知消息（包括直接WS推送和MQ跨实例转发的消息）
    if (msg.type === 'notification') {
      const notif = msg.data as NotificationData
      if (!notif || !notif.type) return

      // 幂等去重：同一ID的消息5分钟内不重复弹窗
      if (this.isDuplicate(notif.id || `${notif.type}-${notif.business_id}`)) return

      if (notif.type === 'new_application') {
        ElNotification({
          title: notif.title || '新的审核申请',
          message: notif.content,
          type: 'info',
          duration: 8000,
          position: 'top-right',
          onClick: () => {
            router.push('/audit')
          }
        })
        this.fetchCounts()
      } else if (notif.type === 'review_result') {
        ElNotification({
          title: notif.title || '审核结果通知',
          message: notif.content,
          type: notif.data?.approved ? 'success' : 'warning',
          duration: 8000,
          position: 'top-right',
          onClick: () => {
            router.push('/my-applications')
          }
        })
        this.fetchCounts()
      } else if (notif.type === 'application_withdrawn') {
        ElNotification({
          title: notif.title || '申请已撤回',
          message: notif.content,
          type: 'warning',
          duration: 6000,
          position: 'top-right'
        })
        this.fetchCounts()
      }

      // 调用注册的自定义处理器
      const handlers = this.handlers.get(notif.type) || []
      const wildcardHandlers = this.handlers.get('*') || []
      ;[...handlers, ...wildcardHandlers].forEach(fn => {
        try { fn(notif) } catch (e) { console.error('[WS] handler error', e) }
      })
      return
    }

    // 未读数更新
    if (msg.type === 'unread_count') {
      if (this.onUnreadCountChange) {
        this.onUnreadCountChange(msg.data?.count ?? 0)
      }
      if (this.onPendingCountChange && msg.data?.pending_count !== undefined) {
        this.onPendingCountChange(msg.data.pending_count)
      }
      return
    }

    // 其他类型消息调用自定义处理器
    const handlers = this.handlers.get(msg.type) || []
    const wildcardHandlers = this.handlers.get('*') || []
    ;[...handlers, ...wildcardHandlers].forEach(fn => {
      try { fn(msg.data) } catch (e) { console.error('[WS] handler error', e) }
    })
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[WS] Max reconnect attempts reached')
      return
    }
    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.min(this.reconnectAttempts, 5)
    console.log(`[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`)
    setTimeout(() => this.doConnect(), delay)
  }

  private startHeartbeat() {
    this.stopHeartbeat()
    this.heartbeatTimer = window.setInterval(() => {
      if (this.ws && this.connected) {
        try {
          this.ws.send('ping')
        } catch (e) {
          console.warn('[WS] Heartbeat failed', e)
        }
      }
    }, 25000)
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer !== null) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  private cleanup() {
    this.stopHeartbeat()
    if (this.ws) {
      this.ws.onopen = null
      this.ws.onmessage = null
      this.ws.onerror = null
      this.ws.onclose = null
      if (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING) {
        this.ws.close()
      }
      this.ws = null
    }
  }

  disconnect() {
    this.intentionalClose = true
    this.connected = false
    this.cleanup()
  }

  isConnected(): boolean {
    return this.connected
  }

  on(event: string, handler: MessageHandler) {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, [])
    }
    this.handlers.get(event)!.push(handler)
  }

  off(event: string, handler?: MessageHandler) {
    if (!handler) {
      this.handlers.delete(event)
      return
    }
    const list = this.handlers.get(event)
    if (list) {
      const idx = list.indexOf(handler)
      if (idx > -1) list.splice(idx, 1)
    }
  }

  setUnreadCountCallback(cb: (count: number) => void) {
    this.onUnreadCountChange = cb
  }

  setPendingCountCallback(cb: (count: number) => void) {
    this.onPendingCountChange = cb
  }

  async fetchCounts() {
    try {
      const { getUnreadCount } = await import('@/api/audit')
      const res = await getUnreadCount()
      if (res.data && this.onUnreadCountChange) {
        this.onUnreadCountChange(res.data.unread_count)
      }
      if (res.data && this.onPendingCountChange) {
        this.onPendingCountChange(res.data.pending_count ?? 0)
      }
    } catch (e) {
      // ignore
    }
  }
}

export const wsService = new WebSocketService()
