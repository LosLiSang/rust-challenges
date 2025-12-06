# Redis 挑战：The One with Redis

## 挑战概述

这是一个 70 分的挑战，要求你从 API 获取 Redis 快照并提取其中的元数据。

## 任务描述

### 背景信息

每次向问题端点发送请求时，你都会获得一个新的、精心制作的 Redis 快照，其中包含以下内容：

- 若干平凡的键值对
- **一个表情符号键** - 你需要处理表情符号
- **一个设有过期时间的键**
- 数据分布在多个数据库中

**特殊情况**：快照头部可能已被篡改，需要注意处理

### 需要提取的数据

你需要从 Redis 快照中提取以下信息：

1. **db_count** - 非空数据库的数量
2. **emoji_key_value** - 表情符号键的值（注意：是键值，不是表情符号的代码点）
3. **expiry_millis** - 唯一设有过期时间的键的时间戳（毫秒）
4. **check_type_of** - 指定键的类型（具体的键名会在问题 JSON 中给出）

## API 端点

### 获取问题

```
GET /challenges/the_redis_one/problem?access_token=...
```

**问题 JSON 结构：**

```json
{
  "rdb": "<base64编码的Redis快照>",
  "requirements": {
    "check_type_of": "<要检查的键名>"
  }
}
```

### 提交解答

```
POST /challenges/the_redis_one/solve?access_token=...
```

**解答 JSON 格式：**

```json
{
  "db_count": <数据库数量>,
  "emoji_key_value": "<表情符号键的值>",
  "expiry_millis": <过期时间戳（毫秒）>,
  "<check_type_of_的实际值>": "<键的类型>"
}
```

#### 重要注意

如果问题 JSON 中包含 `"check_type_of": "overweight_geese"`，则解答 JSON 应该包含 `"overweight_geese": "hash"`（或实际的类型）。

可能的键类型包括：
- string
- list
- set
- hash
- zset（有序集合）
- stream
- 等等

## Redis RDB 格式

RDB 是 Redis 的官方二进制快照格式。你需要编写或使用代码来解析 `.rdb` 文件，提取其中的数据。

### RDB 文件结构基础

- RDB 文件以魔数开头（通常是 "REDIS"）
- 后跟版本号
- 然后是各个选择器部分和数据部分
- 数据部分包含实际的键值对和过期时间信息

## 解决方案思路

1. **解码 base64** - 将 base64 编码的 RDB 转换为二进制数据
2. **解析 RDB 格式** - 读取和理解 RDB 文件结构
3. **提取数据**：
   - 遍历所有数据库
   - 计数非空数据库
   - 找到表情符号键和它的值
   - 找到有过期时间的键和时间戳
   - 检查指定键的类型
4. **提交答案** - 将提取的数据按要求格式提交

## 关键挑战点

- 处理可能被篡改的 RDB 文件头
- 正确理解和解析 RDB 格式
- 处理表情符号字符
- 识别过期时间戳
- 正确获取各种数据类型的类型信息

## 参考资源

- [Redis RDB 格式文档](https://redis.io/topics/protocol-spec)
- [Hackattic 挑战页面](https://hackattic.com/challenges/the_redis_one)
