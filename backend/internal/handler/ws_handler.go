// ws_handler.go 实现 WebSocket 连接升级和实时消息推送处理。
package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	jwtpkg "casbin-demo/pkg/jwt"
	"casbin-demo/pkg/ws"
)

// WsHandler WebSocket 连接处理器。
type WsHandler struct {
	hub *ws.Hub
}

// NewWsHandler 创建 WsHandler 实例。
func NewWsHandler(hub *ws.Hub) *WsHandler {
	return &WsHandler{hub: hub}
}

// websocketUpgrader WebSocket 协议升级器配置。
var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Connect 处理 WebSocket 连接升级请求。
//
// 认证方式：通过 URL query 参数 ?token=<jwt> 传递 Token
// （因为 WebSocket 握手阶段无法设置自定义 Header）。
//
// GET /ws?token=<jwt_token>
func (h *WsHandler) Connect(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		// 也尝试从 Authorization 头获取
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
	}
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未提供认证令牌"})
		return
	}

	claims, err := jwtpkg.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的认证令牌"})
		return
	}

	conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "user_id", claims.UserID, "error", err)
		return
	}

	client := &ws.Client{
		UserID: claims.UserID,
		Roles:  claims.Roles,
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	h.hub.Register(client)
	slog.Info("websocket client connected", "user_id", claims.UserID, "roles", claims.Roles)

	go client.WritePump()
	go client.ReadPump()
}
