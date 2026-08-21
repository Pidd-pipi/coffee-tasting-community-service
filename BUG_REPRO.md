# BUG_REPRO：点赞错误链断裂与计数错位

## Bug 是什么

点赞/取消点赞的错误链断裂：对不存在笔记点赞返回成功；取消未点过的赞返回 500；点赞计数少一。

## 如何触发

服务启动后调用 `POST /api/v1/notes/:id/like`（不存在笔记）或 `DELETE /api/v1/notes/:id/like`（未点赞）。

## 真实错误信息

取消未点过的赞返回 `{"code":50000,"message":"internal server error"}`。
