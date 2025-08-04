package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"game_siteruzou/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan string)
	upgrader  = websocket.Upgrader{}
	mu        sync.Mutex
)

func main() {
	// DB接続
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_NAME"),
	)


	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	q := model.New(conn)

	// WebSocketエンドポイント設定
	http.HandleFunc("/ws", handleConnections) // WebSocketエンドポイント

	// クライアントへの送信ハンドラ
	go handleBroadcast() //通知送信処理

	// Ping監視ループ
	go monitorPing(q)

	// サーバー起動
	fmt.Println("🌐 Listening on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

// WebSocket接続受付
func handleConnections(w http.ResponseWriter, r *http.Request) {

    frontendURL := os.Getenv("FRONTEND_URL")

	allowedOrigins := map[string]bool{
		frontendURL: true,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return allowedOrigins[origin]
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer ws.Close()

	mu.Lock()
	clients[ws] = true
	mu.Unlock()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(clients, ws)
			mu.Unlock()
			break
		}
	}
}

func handleBroadcast() {
	for {
		msg := <-broadcast
		mu.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

// Ping監視処理
func monitorPing(q *model.Queries) {
	ip := os.Getenv("SWITCH_PORT")
	prevStatus := false
	reLoss := regexp.MustCompile(`([0-9.]+)% packet loss`)

	for i := 1; ; i++ {
		cmd := exec.Command("ping", "-c", "1", ip)
		output, _ := cmd.CombinedOutput()
		lines := strings.Split(string(output), "\n")

		for idx, line := range lines {
			if strings.Contains(line, "ping statistics") && idx+1 < len(lines) {
				match := reLoss.FindStringSubmatch(lines[idx+1])
				if len(match) < 2 {
					fmt.Println("ロス率パース失敗")
					break
				}

				var lossRate float64
				fmt.Sscanf(match[1], "%f", &lossRate)
				status := lossRate <= 0.0

				fmt.Printf("----- Ping送信回数: %d -----\n", i)
				fmt.Printf("📊 loss: %.1f%%, 状態: %v\n", lossRate, status)

				if status != prevStatus {

					jst, _ := time.LoadLocation("Asia/Tokyo")
					now := time.Now().In(jst)

					// DB保存
					err := q.InsertPingLog(context.Background(), model.InsertPingLogParams{
						Timestamp: now,
						Status:    status,
						LossRate:  lossRate,
					})
					if err != nil {
						fmt.Println("DB保存エラー:", err)
					} else {
						fmt.Println("✅ DBに保存しました")
					}

					msg := fmt.Sprintf(`{"status":%v,"lossRate":%.1f,"time":"%s"}`,
						status, lossRate, time.Now().Format(time.RFC3339))
					broadcast <- msg
				}

				prevStatus = status
				break
			}
		}

		time.Sleep(5 * time.Second)
	}
}
