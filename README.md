## Ali Middleware 2018 Preliminary Round

### 功能

1. 服务注册与发现 
2. 负载均衡
3. 协议转换

### 流程

1. 启动 etcd 实例
2. 启动三个 Provider 实例，Provider Agent 将 Provider 服务信息写入 etcd 注册中心
3. 启动 Consumer 实例，Consumer Agent 从注册中心中读取 Provider 信息
4. 客户端访问 Consumer 服务
5. Consumer 服务通过 HTTP 协议调用 Consumer Agent
6. Consumer Agent 根据当前的负载情况决定调用哪个 Provider Agent，并使用自定义协议将请求发送给选中的 Provider Agent
7. Provider Agent 收到请求后，将通讯协议转换为 DUBBO，然后调用 Provider 服务
8. Provider 服务将处理后的请求返回给 Agent
9. Provider Agent 收到请求后解析 DUBBO 协议，并将数据取出，以自定义协议返回给 Consumer Agent
10. Consumer Agent 收到请求后解析出结果，再通过 HTTP 协议返回给 Consumer 服务
11. Consumer 服务最后将结果返回给客户端

### 模块

1. 服务的注册和发现，包括服务信息的写入、读取、更新、缓存;(by qzwlecr)
2. 通讯协议的转换，包括自定义协议->DUBBO，DUBBO->自定义协议;(by imzhwk)
3. 通讯协议的转换，包括从HTTP协议中读取相应请求，以HTTP协议发送请求;(by kyle)
4. Consumer的负载均衡。(by imzhwk & qzwlecr)
5. 框架与Docker相关问题。(by qzwlecr)

### 协议细节

TODO: 需实现Provider相关负载信息的实时传递与缓存。

TODO: ZHWK's Implement.

### 合作者

Lichen Zhang(@qzwlecr), Huatian Zhou(@klx3300), Jinyi Xian(@kylerky). 
