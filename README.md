# HERMES


![Version](https://img.shields.io/badge/version-1.0.0-green.svg)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](http://opensource.org/licenses/MIT)

## Overview

- 基于RSA用于项目之间交互的中间件
- 1.0.0开发完成
- 后续开发Java SDK

## Configuration

- 配置文件需放置于与可执行文件同一目录并命名为hermes.json

#### Example

```json
{
  "port": 8080,
  "timeout": 180,
  "keyLength": 1024,
  "backupPath": "/root/backup/"
}
```
1. port：服务启动端口号

1. timeout:：服务过期时间，单位秒

1. keyLength：密钥长度，加密内容不能超过长度-11

1. backupPath：服务信息备份地址
