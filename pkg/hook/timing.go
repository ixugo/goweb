package hook

import (
	"log/slog"
	"runtime"
	"time"
)

// UseTiming 计算函数花销，超过 limit 用 error 级别记录
// cost := UseTiming(time.Second)
// defer cost()
// 业务操作
func UseTiming(limit time.Duration) func() {
	now := time.Now()
	return func() {
		sub := time.Since(now)
		pc, _, _, _ := runtime.Caller(1)
		fn := runtime.FuncForPC(pc)

		log := slog.With("cost", sub, "caller", fn.Name())
		if sub > limit {
			log.Error("timing")
		} else {
			log.Debug("timing")
		}
	}
}
