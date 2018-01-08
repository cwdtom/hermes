# Hermes


![Version](https://img.shields.io/badge/version-2.1.0-green.svg)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](http://opensource.org/licenses/MIT)

## Overview

- 基于RSA用于项目之间交互的中间件
- Java端SDK (https://github.com/cwdtom/hermes-java)

## Configuration

- 配置文件需放置于与可执行文件同一目录并命名为hermes.json

- Example (hermes.json)
    ```json
    {
      "port": 8080,
      "timeout": 180,
      "keyLength": 1024,
      "backupPath": "/root/backup/",
      "password": "123456",
      "whiteList": [
        "127.0.0.1"
      ]
    }
    ```

1. port：服务启动端口号

1. timeout:：服务过期时间，单位秒

1. keyLength：密钥长度，加密内容不能超过长度-11

1. backupPath：服务信息备份地址

1. password: 监控页面登录密码

1. whiteList: ip白名单，不在白名单的ip不允许注册，不填写或留空表示允许所有ip（生产环境请勿留空）

## Usage

1. http://your_host:port/ 可以登录监控页面
