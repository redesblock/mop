# 软件安装
安装启动MOP，您需要完成以下过程。
1. 安装客户端。
2. 配置并启动客户端。
3. 提供币安智能区块链(Binance Smart Chain)的BNB和MOP资金。
4. 等待部署支票簿、同步voucher数据。
5. 检查客户端正在运行。

### 1. 安装客户端

- Ubuntu

```shell
wget https://github.com/redesblock/mop/releases/download/v0.9.1/mop_0.9.1_amd64.deb
sudo dpkg -i mop_0.9.1_amd64.deb
```

- CentOS

```shell
wget https://github.com/redesblock/mop/releases/download/v0.9.1/mop_0.9.1_amd64.rpm
sudo rpm -i mop_0.9.1_amd64.rpm
```

-  MacOS

```shell
brew tap redesblock/tap
brew install mop
```

如果您的系统不受支持，您可以选择其他方式进行安装。

- [脚本安装](installation/install-from-script.md)

- [源码安装](installation/install-from-source.md)

### 2. 配置并启动客户端

在第一次启动 MOP之前，您需要对其配置文件`mop.yaml`进行配置以满足您的需要。

- Ubuntu

```shell
sudo vi /etc/mop/mop.yaml
sudo systemctl restart mop
```

- Centos

```shell
sudo vi /etc/mop/mop.yaml
sudo systemctl restart mop
```

- MacOS

```shell
vi /usr/local/etc/mop/mop.yaml
brew services restart mop
```

此外，您也可以支持使用命令行配置项进行配置并启动

```shell script
mop start --full-node --db-open-files-limit 10000 --network-id 97 --mainnet=false --verbosity 4 --debug-api-enable --cors-allowed-origins "*" --bsc-rpc-endpoint http://202.83.246.155:8575 --bsc-rpc-endpoint https://nd-809-271-506.p2pify.com/f3fc360ea84aa05b0ad489dc9f7618e6 --bootnode /ip4/202.83.246.155/tcp/1684/p2p/16Uiu2HAmPqr2vmnwZi6HhTmWoCEVx2pD37m3p9G5dfNYCMrormLf --password "12345678"
```

### 3. 提供币安智能区块链(Binance Smart Chain)的BNB和MOP资金
客户端必须部署一个支票簿合约来跟踪它与其他客户端的数据转发交易。部署合约需要BNB和MOP资金。

首先，通过找出你的币安智能链地址地址。

```shell
sudo mop-get-addr
```

 另外，您也可以通过日志文件查找你的币安智能链地址地址。

```shell
cat ~/.mop/logs/mop.log | grep "using BNB Smart Chain" | head 1
```

确定MOP币安智能链地址后，使用BNB和MOP为您的节点提供资金。如果时间过长，此时您可能需要重新启动节点。

### 4. 等待部署支票簿、同步voucher数据

首次启动时，客户端必须将支票簿部署到币安智能区块链(Binance Smart Chain)，并同步voucher数据，以便它可以在存储或转发时检查块的有效性。

这可能需要一段时间，所以请耐心等待！完成后，您将看到开始添加对等点并连接到网络。

您可以在这发生时通过查看日志来关注进度。

```shell
tail -f  ~/.mop/logs/mop.log
```

### 5. 检查客户端正在运行
MOP客户端获得资金、部署支票簿、同步存储后, HTTP API开始监听端口1683。
可通过向主机端口1683发送GET请求，检查一切是否按预期工作。
```shell script
curl localhost:1633
```

