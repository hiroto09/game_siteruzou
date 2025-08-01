package main

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "os/exec"
    "regexp"
    "strings"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "game_siteruzou/model"
)

func main() {
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
                    err := q.InsertPingLog(context.Background(), model.InsertPingLogParams{
                        Timestamp: time.Now(),
                        Status:    status,
                        LossRate:  lossRate,
                    })
                    if err != nil {
                        fmt.Println("DB保存エラー:", err)
                    } else {
                        fmt.Println("✅ DBに保存しました")
                    }
                }

                prevStatus = status
                break
            }
        }

        time.Sleep(5 * time.Second)
    }
}
