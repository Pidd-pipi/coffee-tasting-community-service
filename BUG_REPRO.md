# BUG_REPRO：限流器并发竞态与计数错乱

## Bug 是什么

`internal/middleware/rate_limiter.go` 的限流器存在多处并发/逻辑缺陷：

1. 读写 `limits` map 时没有加锁，高并发下触发 data race；
2. 窗口复位判断用了反转的条件，窗口到期后计数不复位；
3. 每次请求会把计数加两次，导致第一条请求就可能被限；
4. 默认限流配置（RATE_LIMIT_REQUESTS / RATE_LIMIT_WINDOW_SECONDS）解析错误。

## 如何触发

服务启动后（`docker compose up -d --build`），对带限流的接口（如 `POST /api/v1/users/login`）快速连续请求即可观察到异常计数；用 `-race` 跑并发测试会直接报 data race。

定向复现：

- `go test -race ./internal/rateqc -run '^TestR2ConcurrentRequestsNoRace$' -count=1`（红）

## 真实错误信息

并发测试输出：

```
WARNING: DATA RACE
Read at 0x00c0001efa70 by goroutine 13:
      internal/middleware/rate_limiter.go:37
Previous write at 0x00c0001efa70 by goroutine 8:
      internal/middleware/rate_limiter.go:40
```
