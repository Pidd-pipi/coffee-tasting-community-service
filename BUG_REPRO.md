# BUG_REPRO：个人主页统计 nil 与聚合报错

## Bug 是什么

无笔记用户的主页统计字段返回 null 而非空数组/0，平均分聚合在无数据时扫描 NULL 报错，note_count 错位。

## 如何触发

服务启动后访问 `GET /api/v1/users/:id/profile`（一个没有发过笔记的用户）。

## 真实错误信息

`top_origins` 与 `notes` 为 null，`avg_score` 计算报 `Scan error ... NULL to float64`。
