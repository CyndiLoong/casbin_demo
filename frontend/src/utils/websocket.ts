// WebSocket 服务：管理实时连接、自动重连、消息分发
import { ElNotification } from 'element-plus'
import { useUserStore } from '@/store/user'

type MessageHandler = (data: any) => void

interface WSMessage {
  type: string
  target_type: string
  target_id?: number
  room?: string
  data: any
  timestamp: string
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
      console.error('WebSocket creation failed', e)
      this.scheduleReconnect()
      return
    }

    this.ws.onopen = () => {
      console.log('[WS] Connected')
      this.connected = true
      this.reconnectAttempts = 0
      this.startHeartbeat()
      // 连接成功后拉取一次未读/待审数量
      this.fetchCounts()
    }

    this.ws.onmessage = (event) => {
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

  private handleMessage(msg: WSMessage) {
    const data = msg.data

    if (data.type === 'new_application') {
      ElNotification({
        title: data.title || '新的审核申请',
        message: data.content,
        type: 'info',
        duration: 8000,
        onClick: () => {
          window.location.hash = '#/audit'
        }
      })
      this.fetchCounts()
    } else if (data.type === 'review_result') {
      ElNotification({
        title: data.title || '审核结果通知',
        message: data.content,
        type: data.data?.approved ? 'success' : 'warning',
        duration: 8000,
        onClick: () => {
          window.location.hash = '#/my-applications'
        }
      })
      this.fetchCounts()
    } else if (data.type === 'unread_count') {
      if (this.onUnreadCountChange) {
        this.onUnreadCountChange(data.count)
      }
    }

    // 调用注册的自定义处理器
    const handlers = this.handlers.get(data.type) || []
    const wildcardHandlers = this.handlers.get('*') || []
    ;[...handlers, ...wildcardHandlers].forEach(fn => {
      try { fn(data) } catch (e) { console.error('[WS] handler error', e) }
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
        this.onPendingCountChange(res.data.pending_count)
      }
    } catch (e) {
      // ignore
    }
  }
}

export const wsService = new WebSocketService()
