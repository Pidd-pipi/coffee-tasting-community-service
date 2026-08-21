# BUG_REPRO：豆种编辑零值覆盖与 nil 解引用

## Bug 是什么

豆种编辑接口存在零值覆盖与错误传播问题：留空字段会覆盖数据库已有值；处理法为空时不校验；更新不存在的豆种会触发 nil 指针解引用 panic。

## 如何触发

服务启动后调用 `PUT /api/v1/beans/:id`（admin），传入只有部分字段的 JSON；或对不存在的 id 执行更新。

## 真实错误信息

```
panic: runtime error: invalid memory address or nil pointer dereference
      internal/service/bean_service.go:52
```
