# table-demo

## Run demo

- 创建数据库，执行 sql/\*
- 修改 app.yml 的数据库配置
- 编译

  ```bash
  go build -tags netgo -o app/app ./app
  docker-compose build
  ```

- 前端文件(dist)放置到 nginx/html/下
- 启动程序

  ```bash
  docker-compose up
  ```

- Visit

  ```bash
  http://localhost:8000/table?tableID=tables.MemTable&tableName=%E5%86%85%E5%AD%98%E8%A1%A8%E6%A0%BC%E6%B5%8B%E8%AF%95
  http://localhost:8000/table?tableID=tables.Testtable&tableName=%E6%95%B0%E6%8D%AE%E5%BA%93%E8%A1%A8%E6%A0%BC&enableOperation=true

  ```
