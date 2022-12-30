# Uploder 

根据设置的监控目录列表，每隔一段时间将符合条件的子目录上传到mop网络中，并通过子目录的相对路径查询MOP网络访问地址。

> 符合条件的子目录：
>
> - 子目录可以找到索引文件 (如.m3u8文件)
> - 索引文件的更新时间应在一个轮询间隔(如10分钟)之前。
>
> 程序运行在数据资源所在机器，监控目录为程序运行机器的文件路径

使用说明：

```shell
~$ mop-uploader -h
automatically upload files or folders to mop cluster

Usage:
  mop-uploader [command]

Available Commands:
  addVoucher    add an usable voucher id from special node for file uploading
  addWatch      add the folder path or file path for monitoring upload
  completion    Generate the autocompletion script for the specified shell
  getUpload     get mop gateway url
  help          Help about any command
  listUpload    list uploaded folders
  listVoucher   list vouchers for uploading file
  listWatch     list folder path or file path for monitoring upload
  removeVoucher remove the voucher for uploading
  removeWatch   remove the folder path or file path for monitoring upload
  serve         automatically upload files or folders to mop cluster

Flags:
      --config string          config file (default is config.yaml)
```

## 命令行

### 启动服务`serve`

```shell
~$ mop-uploader serve -h
automatically upload files or folders to mop cluster

Usage:
  mop-uploader serve [flags]

Flags:
      --gateway string    mop gateway (default "https://gateway.mopweb3.com:13443")
  -h, --help              help for serve
      --interval string   watcher poll interval (default "10m")
      --port string       listen port (default "8082")
      --database_dsn string    database source name (default "sqlite.db")
      --database_mode string   database mode, sqlite、mysql、postgres (default "sqlite")

Global Flags:
      --config string          config file (default is config.yaml)
```

参数：

- port： 服务监听端口，默认8082。
- interval： 轮询查询监控目录文件列表，查找合适的文件并上传文件，默认10分钟。
- gateway： mop cluster gateway域名，在获取文件访问地址时自动添加，默认https://gateway.mopweb3.com:13443
- database_mode： 数据库类型，默认是sqlite数据库。 
- database_dsn： 数据库访问路径。默认是启动目录下sqlite.db文件

HTTP接口：

http://127.0.0.1:8082/swagger/index.html

### 新建监控目录`addWatch`

```shell
~$ mop-uploader addWatch -h
add the folder path or file path for monitoring upload

Usage:
  mop-uploader addWatch path index_ext [flags]

Flags:
  -h, --help          help for addWatch
      --node string   node api (default "http://127.0.0.1:8082")

Global Flags:
      --config string          config file (default is config.yaml)
```

参数：

- path： 监控目录路径
- index_ext： 索引文件后缀 (如.m3u8)

- node： mop-uploader服务地址，默认http://127.0.0.1:8082。

### 移除监控目录`removeWatch`

```shell
~$ mop-uploader removeWatch -h
remove the folder path or file path for monitoring upload

Usage:
  mop-uploader removeWatch path [flags]

Flags:
  -h, --help          help for removeWatch
      --node string   node api (default "http://127.0.0.1:8082")

Global Flags:
      --config string          config file (default is config.yaml)
```

- path： 监控目录路径

- node： mop-uploader服务地址，默认http://127.0.0.1:8082。

### 查看监控目录列表`listWatch`

```shell
~$ mop-uploader listWatch -h
list folder path or file path for monitoring upload

Usage:
  mop-uploader listWatch [page_size] [page_num] [flags]

Flags:
  -h, --help          help for listWatch
      --node string   node api (default "http://127.0.0.1:8082")

Global Flags:
      --config string          config file (default is config.yaml)
```

参数：

- node： mop-uploader服务地址，默认http://127.0.0.1:8082。

### 查看上传文件列表`listUpload`

```shell
~$ mop-uploader listUpload -h
list uploaded folders

Usage:
  mop-uploader listUpload [page_size] [page_num] [flags]

Flags:
  -h, --help          help for listUpload
      --node string   node api (default "http://127.0.0.1:8082")

Global Flags:
      --config string          config file (default is config.yaml)
```

参数：

- node： mop-uploader服务地址，默认http://127.0.0.1:8082。

### 获取上传文件访问地址`getUpload`

```shell
~$ mop-uploader getUpload -h
get mop gateway url

Usage:
  mop-uploader getUpload path [flags]

Flags:
  -h, --help          help for getUpload
      --node string   node api (default "http://127.0.0.1:8082")

Global Flags:
      --config string          config file (default is config.yaml)
```

- path： 相对监控目录的相对路径
- node： mop-uploader服务地址，默认http://127.0.0.1:8082。

### 新增Voucher`addVoucher`

### 移除Voucher`removeVoucher`

### 查看Voucher列表`listVoucher`