# polaris-cleanup

[English Document](./README.md)

Polaris-cleanup 作为守护进程运行，用于清理 Polaris 中注册的无用资源

您可以在 Polaris 集群上应用以下清理工作:

- 清理软删除服务、服务实例、服务治理规则
- 清理长期不健康的服务实例
- 清理非控制台手动创建、无实例、无服务治理规则的服务

## 配置文件

```yaml
# 北极星server存储层链接信息
store:
  dbHost: ##DBHOST##
  dbPort: ##DBPORT##
  dbName: ##DBNAME##
  dbUser: ##DBUSER##
  dbPwd: ##DBPWD##
# 北极星server的接入点信息，这里连接的是北极星的http-apiserver
server:
  # 支持配置多个北极星接入点
  endpoints:
    - 127.0.0.1:8090
  # 如果北极星server开启了鉴权，则需要填写用户/用户组的token凭据
  authToken: 
  # 请求ID前缀
  requestPrefix: polaris-cleanup-
# 数据清理相关控制变量
cleanup:
  # 数据需要多久之后才能认为是无效数据
  deleteLimitedTime:
  # 限制删除的总数量，控制从DB中获取的数据量，避免对DB增加负载
  deleteLimitedNum:
  # 每次批删除的数量
  batchDeleteNum:
# 要开启的任务类型
openJob:
  # 清理软删除的服务实例
  - DeleteSoftDeleteInstance
  # 清理长期不健康的实例
  - DeleteUnHealthyInstance
```