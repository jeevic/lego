#lego  - happy use go 
go脚手架 封装常用组件 支持http grpc, 定时任务, 守护态程序开发

#### 特性
- 封装常用组件,降低开发使用成本
- 集成viper配置管理
- 集成logrus日志管理, 支持多日志配置, 支持自定义日志格式 便于业务定制
- 集成gin 做http server 支持令牌桶限流
- 集成grpc server 集成日志记录 限流 recover keepalive拦截器功能
- 集成 gocron 定时任务调度 支持秒级别定时和指定时间定时
- 集成 grpc客户端 支持连接池模式 提升并发性能
- 集成 redis, codis(自开发) redis 客户端 
- 集成 zookeeper 客户端
- 集成 mongo 客户端
- 集成 httplib(来源beego) http请求组件
- 集成 swagger ui
- 接管信号 支持http grpc graceful shutdown
- 集成 dingding robot机器人(自开发), 安全加签模式 支持发送text、link、markdown 类型消息
- 脚手架核心只依赖配置管理, 日志, gin, grpc 模块, 包尽量小, 其他模块以组件形式提供

项目使用demo参考
参考脚手架使用demo: https://github.com/jeevic/lego-demo