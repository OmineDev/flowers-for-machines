# 日志服务器 & 标准 HTTP 接口文档





## 基本信息
> 协议版本: v1.0.0  
> 最后更改: 11th July, 2025 





## 目录
- [日志服务器 \& 标准 HTTP 接口文档](#日志服务器--标准-http-接口文档)
  - [基本信息](#基本信息)
  - [目录](#目录)
  - [HelloWorld](#helloworld)
    - [基本信息](#基本信息-1)
    - [返回值](#返回值)
  - [LogRecord](#logrecord)
    - [描述](#描述)
    - [基本信息](#基本信息-2)
    - [枚举值](#枚举值)
    - [请求表单](#请求表单)
    - [返回表单](#返回表单)
  - [LogReview](#logreview)
    - [描述](#描述-1)
    - [基本信息](#基本信息-3)
    - [请求表单](#请求表单-1)
    - [返回表单](#返回表单-1)
    - [结构体说明](#结构体说明)
  - [LogFinishReview](#logfinishreview)
    - [描述](#描述-2)
    - [基本信息](#基本信息-4)
    - [请求表单](#请求表单-2)
    - [返回表单](#返回表单-2)
  - [SetAuthKey](#setauthkey)
    - [描述](#描述-3)
    - [基本信息](#基本信息-5)
    - [请求表单](#请求表单-3)
    - [返回表单](#返回表单-3)





## HelloWorld
### 基本信息
| 项          | 值                                  |
| ----------- | ----------------------------------- |
| Method      | GET                                 |
| URL         | https://log-record.eulogist-api.icu |
| ContentType | -                                   |
| Response    | text/plain                          |
 

### 返回值
```
Hello, World!
```





## LogRecord
### 描述
向日志服务器发送一个日志。
在设计上，应该只在出现问题时发送日志。

### 基本信息
| 项          | 值                                             |
| ----------- | ---------------------------------------------- |
| Method      | POST                                           |
| URL         | https://log-record.eulogist-api.icu/log_record |
| ContentType | application/json                               |
| Response    | JSON                                           |

### 枚举值
- Source (字符串)
    - OmegaBuilder
    - ToolDelta
    - FunOnBuilder
    - YsCloud
- System Name (字符串)
    - ChangeConsolePosition
    - PlaceNBTBlock
    - PlaceLargeChest
    - GetNBTBlockHash

### 请求表单
| 键               | 值类型 | 值描述                                                                             |
| ---------------- | ------ | ---------------------------------------------------------------------------------- |
| source           | 字符串 | 日志来源，可能是 **ToolDelta**, **OmegaBuilder**, 等等 (见上方的枚举值)            |
| user_name        | 字符串 | 用户名，例如用户在 **ToolDelta** 面板的名称                                        |
| bot_name         | 字符串 | 用户所使用的导入 NBT 方块的机器人的名称                                            |
| create_unix_time | 整数   | 日志创建的时间戳                                                                   |
| system_name      | 字符串 | 产生日志的系统名，如 **PlaceNBTBlock**, **PlaceLargeChest**, 等等 (见上方的枚举值) |
| user_request     | 字符串 | 用户的原始请求的 JSON 字符串                                                       |
| error_info       | 字符串 | 在执行用户的原始请求时出现的错误信息                                               |

### 返回表单
| 键            | 值类型 | 值描述                                            |
| ------------- | ------ | ------------------------------------------------- |
| success       | 布尔值 | 请求是否成功处理                                  |
| error_info    | 字符串 | 如果请求处理失败，则这个字段指示具体的错误信息    |
| log_unique_id | 字符串 | 如果请求处理成功，则这个字段是所提交日志的唯一 ID |





## LogReview
### 描述
从日志服务器上检索日志。
仅限已被授权的管理员使用。

### 基本信息
| 项          | 值                                             |
| ----------- | ---------------------------------------------- |
| Method      | POST                                           |
| URL         | https://log-record.eulogist-api.icu/log_review |
| ContentType | application/json                               |
| Response    | JSON                                           |
 
### 请求表单
| 键               | 值类型     | 值描述                                                                                                                 |
| ---------------- | ---------- | ---------------------------------------------------------------------------------------------------------------------- |
| auth_key         | 字符串     | 管理员的令牌                                                                                                           |
| include_finished | 布尔值     | 要检索的日志是否包含那些已被审阅完成 (已被标记为已处理) 的日志                                                         |
| source           | 字符串列表 | 如果非空，则只检索 `日志来源` 在这个字符串列表内的日志                                                                 |
| log_unique_id    | 字符串列表 | 如果非空，则只检索 `日志唯一 ID` 在这个字符串列表内的日志                                                              |
| user_name        | 字符串列表 | 如果非空，则只检索 `用户名` 在这个字符串列表内的日志                                                                   |
| bot_name         | 字符串列表 | 如果非空，则只检索 `机器人名称` 在这个字符串列表内的日志                                                               |
| start_unix_time  | 整数       | 表示时间戳。如果它和 `end_unix_time` 都非 0，则检索产生时间在 `start_unix_time` 到 `end_unix_time` 内的日志 (含边界)   |
| end_unix_time    | 整数       | 表示时间戳。如果它和 `start_unix_time` 都非 0，则检索产生时间在 `start_unix_time` 到 `end_unix_time` 内的日志 (含边界) |
| system_name      | 字符串列表 | 如果非空，则只检索 `系统名` 在这个字符串列表内的日志                                                                   |

### 返回表单
| 键          | 值类型     | 值描述                                         |
| ----------- | ---------- | ---------------------------------------------- |
| success     | 布尔值     | 请求是否成功处理                               |
| error_info  | 字符串     | 如果请求处理失败，则这个字段指示具体的错误信息 |
| log_records | 字符串列表 | 如果请求处理成功，则这个列表包含检索到的日志   |

### 结构体说明
`log_records` 中的每个日志满足下面的结构体。

| 键               | 值类型 | 值描述                                                                |
| ---------------- | ------ | --------------------------------------------------------------------- |
| log_unique_id    | 字符串 | 这个日志的唯一 ID                                                     |
| review_states    | 整数   | 这个日志的审阅状态。为 0 指示未完成审阅；为 1 表示已完成审阅 (已处理) |
| source           | 字符串 | 这个日志的 `日志来源`                                                 |
| user_name        | 字符串 | 这个日志中记录的 `用户名`                                             |
| bot_name         | 字符串 | 这个日志中记录的 `机器人名称`                                         |
| create_unix_time | 整数   | 这个日志创建的时间戳                                                  |
| system_name      | 字符串 | 这个日志产自的 `系统名`                                               |
| user_request     | 字符串 | 用户原始请求的 JSON 数据                                              |
| error_info       | 字符串 | 用户原始请求后得到的错误信息                                          |





## LogFinishReview
### 描述
将服务器上的指定日志标记为已被审阅 (已被处理)。
仅限已被授权的管理员使用。

### 基本信息
| 项          | 值                                                    |
| ----------- | ----------------------------------------------------- |
| Method      | POST                                                  |
| URL         | https://log-record.eulogist-api.icu/log_finish_review |
| ContentType | application/json                                      |
| Response    | JSON                                                  |
 
### 请求表单
| 键            | 值类型     | 值描述                                      |
| ------------- | ---------- | ------------------------------------------- |
| auth_key      | 字符串     | 管理员的令牌                                |
| log_unique_id | 字符串列表 | 要标记为已被审阅 (已被处理) 的日志的唯一 ID |

### 返回表单
| 键         | 值类型 | 值描述                                         |
| ---------- | ------ | ---------------------------------------------- |
| success    | 布尔值 | 请求是否成功处理                               |
| error_info | 字符串 | 如果请求处理失败，则这个字段指示具体的错误信息 |





## SetAuthKey
### 描述
新增或删除管理员令牌。
仅限已被授权的管理员使用。

### 基本信息
| 项          | 值                                               |
| ----------- | ------------------------------------------------ |
| Method      | POST                                             |
| URL         | https://log-record.eulogist-api.icu/set_auth_key |
| ContentType | application/json                                 |
| Response    | JSON                                             |
 
### 请求表单
| 键              | 值类型 | 值描述                                                                 |
| --------------- | ------ | ---------------------------------------------------------------------- |
| token           | 字符串 | 管理员的令牌                                                           |
| auth_key_action | 整数   | 要进行的操作。为 0 指示新增一个管理员令牌，为 1 指示删除一个管理员令牌 |
| auth_key_to_set | 字符串 | 要新增或删除的管理员令牌                                               |

### 返回表单
| 键         | 值类型 | 值描述                                         |
| ---------- | ------ | ---------------------------------------------- |
| success    | 布尔值 | 请求是否成功处理                               |
| error_info | 字符串 | 如果请求处理失败，则这个字段指示具体的错误信息 |