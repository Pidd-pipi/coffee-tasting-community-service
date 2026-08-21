# BUG_REPRO：评论删除/创建错误链断裂

## Bug 是什么

评论模块的错误处理存在多处错误链断裂，导致：

1. 删除不存在的评论返回 500，而不是 404；
2. 在不存在（或已删除）的笔记下发表评论能成功，而不是 404；
3. 非作者删除评论不会返回 403；
4. 被 `fmt.Errorf("...: %w", appErr)` 包装的 `AppError` 会丢失状态码，变成 500。

## 如何触发

服务启动后（`docker compose up -d --build`，后端监听 `:29601`）：

- `DELETE /api/v1/comments/:id`，传入一个不存在的评论 id → 返回 500；
- `POST /api/v1/notes/:id/comments`，传入一个不存在的笔记 id → 返回 201；
- 用非作者账号删除他人评论 → 返回 200。

也可以直接跑定向测试复现：

- `go test ./internal/commentqc -run '^TestC1DeleteMissingCommentIs404$' -count=1`（红）

## 真实错误信息

删除不存在评论时，接口返回：

```json
{"code":50000,"message":"internal server error"}
```

定位时可见错误链被 `%v` 打断，`errors.Is(err, repository.ErrNotFound)` 永远为 false。
