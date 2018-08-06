### keepalived说明

- `master_keepalived.conf`是主broker所在keepalived环境所需配置；需改名移动到`/etc/keepalived/`目录下
- `backup_keepalived.conf`是备broker所在keepalived环境所需配置；需改名移动到`/etc/keepalived/`目录下
- `broker_detect.sh`是检测broker是否正常服务的脚本，可在其中添加ping网关操作，用于防脑裂；需移动到`/etc/keepalived/`目录下
- `broker_detect.go`是检测broker是否正常服务的程序，运行`go build`编译生成可执行文件，由broker_detect.sh调用，需移动到`/etc/keepalived/`目录下
- `restart_service.sh`是服务挂掉后（非掉电），尝试周期性自动重启服务的脚本，建议作为crontab任务
