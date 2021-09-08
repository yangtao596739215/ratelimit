package main

import (
	"fmt"
	"sync"
	"time"
)

type TokenRatelimit struct {
	burst  int64   // 桶的容量
	limit  float64 // 每秒的补充速率
	tokens int64   // 桶里目前令牌的数量

	mu   sync.Mutex
	last time.Time // 上一次消耗令牌的时间
}

// 刷新桶的数量，根据场景,可以在每次获取的时候刷新，也可以启一个goroutine定时刷新，但是这样的话就要考虑多协程安全问题
func (lim *TokenRatelimit) refreshToken(now time.Time) {

	lim.tokens += int64(now.Sub(lim.last).Seconds() * lim.limit)
	if lim.tokens > lim.burst {
		lim.tokens = lim.burst
	}
}

func (lim *TokenRatelimit) Allow() bool {
	return lim.AllowN(1)
}

func (lim *TokenRatelimit) AllowN(n int64) bool {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	now := time.Now()
	lim.refreshToken(now)
	if lim.tokens < n {
		return false
	}
	lim.tokens -= n
	lim.last = now
	return true
}

// 构造函数不用传tokens和时间，因为refresh的时候，根据时间0值，第一次会直接把桶填满
func NewTokenRatelimit(burst int64, limit float64) *TokenRatelimit {
	return &TokenRatelimit{burst: burst, limit: limit}
}

func main() {
	limiter := NewTokenRatelimit(5, 3) // 容量为5，每秒补充三个
	for {
		n := 4 // 每秒取四个
		for i := 0; i < n; i++ {
			go func(i int) {
				if !limiter.Allow() {
					fmt.Println("forbid:", i)
				}
				fmt.Println("allow", i)
			}(i)
		}
		time.Sleep(time.Second)
	}
}

// 结果说明，一开始桶里有5个，每秒加3个取4个，第一次和第二次可以都取到。从地三次开始，每次都只能取到三个
