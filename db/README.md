唯一索引

`db.configures.ensureIndex({"serviceKey":1}, {"unique":true})`


导出

`mongodump -d sona-configures -o .`

导入

`mongorestore -d sona-configures sona-configures`