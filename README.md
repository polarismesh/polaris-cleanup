# polaris-cleanup

[中文文档](./README-zh.md)

polaris-cleanup runs as a daemon to cleanup the useless resources registered in polaris

Here are 3 cleanups you can apply on your polaris cluster:

- clean up the soft deleted service, service instance, service governance rules
- clean up long term unhealthy service instance
- clean up the non console manual creation, without service instance, no service governance rules

## Configuration

```yaml
# polaris server storage layer link information
store:
  dbHost: ##DBHOST##
  dbPort: ##DBPORT##
  dbName: ##DBNAME##
  dbUser: ##DBUSER##
  dbPwd: ##DBPWD##
# polaris server access point information, here is the HTTP-APISERVER of the polaris server
server:
  # support configuration of multiple polaris server access points
  endpoints:
    - 127.0.0.1:8090
  # If the polaris server starts the verification, you need to fill in the tokens of the user/user group
  authToken: 
  # Request ID prefix
  requestPrefix: polaris-cleanup-
# Data cleaning related control variables
cleanup:
  # How long does it take for data to be considered as invalid data
  deleteLimitedTime:
  # Limit the total amount of delete, control the amount of data obtained from DB, and avoid adding loads to DB
  deleteLimitedNum:
  # Number of deletion each time
  batchDeleteNum:
# Type of task to open
openJob:
  # Clean up the service instance of soft deletion
  - DeleteSoftDeleteInstance
  # Clean up long term unhealthy instance
  - DeleteUnHealthyInstance
```