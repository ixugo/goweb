package data

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/ixugo/goweb/internal/conf"
)

func TestAAbc(t *testing.T) {
	db, err := SetupDB(&conf.Bootstrap{
		Data: conf.Data{
			Database: conf.Database{
				Dsn: "postgres://postgres:5307c6bd4b12_c@localhost:7789/sitong?sslmode=disable",
			},
		},
	}, slog.Default())
	if err != nil {
		panic(err)
	}

	const limit = 100
	i := 360
	// ch := make(chan struct{}, 10)
	// for {
	ids := make([]int, 0, 10)
	if err := db.Raw(`SELECT id FROM users WHERE last_login_info->>'version'='0.6.0' LIMIT 100`).Scan(&ids).Error; err != nil {
		fmt.Println(err)
		return
	}
	i++

	for _, id := range ids {
		if id <= 0 {
			continue
		}
		// ch <- struct{}{}
		// go func(id int) {
		// defer func() {
		// <-ch
		// }()
		fmt.Println("修复 UID:", id)
		var data map[string]any
		if err := db.Raw(`SELECT misc from logs where act_type='login' AND uid=? ORDER BY id DESC  LIMIT 1`, id).Scan(&data).Error; err != nil {
			t.Error(err)
			panic(err)
		}
		out, ok := data["misc"]
		if !ok {
			fmt.Println("uid:", id, "无登录记录")
			continue
		}

		if err := db.Debug().Exec(`UPDATE users SET last_login_info=?::jsonb WHERE id=?`, out, id).Error; err != nil {
			fmt.Println(err, "id:", id)
		}
		// }(id)

	}
	time.Sleep(10 * time.Second)
	fmt.Println("end")
	// }
}
