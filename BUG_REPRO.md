# BUG_REPRO：配方列表切片共享污染

## Bug 是什么

配方列表在查询后做了原地过滤（`items[:0]` 复用底层数组），导致步骤数据串场、total/page_size 错位、详情步骤被截断。

## 如何触发

服务启动后调用 `GET /api/v1/recipes?page=1&page_size=10`，观察列表中的 steps 字段是否互相串场。

## 真实错误信息

列表返回的某些配方 steps 与数据库内容不一致，翻页返回的 total 与 page_size 不匹配。
