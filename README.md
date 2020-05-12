### 根据阿里云提供的RDS慢查询api，将慢查询记录以邮件形式发送给相关人员

---
###  部署流程
1. mac下执行package-mac.sh，将编译生成linux下可执行文件
2. windows下执行package-windows.sh，将编译生成linux下可执行文件
3. 将可执行文件rds4go、配置文件config.ini(修改相关配置)、邮件模板文件email.html上传至服务器某文件夹下
4. chmod a+x rds4go 
5. 定时任务crontab（比如每天早晨8点发送前一天的慢查询日志)：<br/>
```0 8 * * * sh /home/netcafe/script/rds/slowlog.sh```
