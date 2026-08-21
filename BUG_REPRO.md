# BUG_REPRO：上传句柄生命周期错位

## Bug 是什么

文件上传逻辑中文件句柄被提前关闭，扩展名/尺寸校验错位，错误状态码误映射为 500。

## 如何触发

服务启动后调用 `POST /api/v1/uploads` 上传图片（合法 png/webp、10KB 左右）。

## 真实错误信息

```
write ...: file already closed
```
