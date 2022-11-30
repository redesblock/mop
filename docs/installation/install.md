# 软件安装
更多人通过运行MOP节点为去中心化系统的健康和分布式特性做出贡献时，同时发挥系统最佳作用。

您需要完成以下过程，完成运行
1. 安装客户端。
2. 启动客户端。
3. 提供币安智能区块链(Binance Smart Chain)的BNB和MOP资金。
4. 等待部署支票簿、同步voucher数据。
5. 检查客户端正在运行。

### 1. 安装客户端

- [脚本安装](installation/install-from-script.md)
- [源码安装](installation/install-from-source.md)

### 2. 启动客户端

```shell script
mop start --full-node --db-open-files-limit 10000 --network-id 97 --mainnet=false --verbosity 4 --debug-api-enable --cors-allowed-origins "*" --bsc-rpc-endpoint http://202.83.246.155:8575 --bsc-rpc-endpoint https://nd-809-271-506.p2pify.com/f3fc360ea84aa05b0ad489dc9f7618e6 --bootnode /ip4/202.83.246.155/tcp/1684/p2p/16Uiu2HAmPqr2vmnwZi6HhTmWoCEVx2pD37m3p9G5dfNYCMrormLf --data-dir /sd/data --password "123456"
```

### 3. 提供币安智能区块链(Binance Smart Chain)的BNB和MOP资金
客户端必须部署一个支票簿合约来跟踪它与其他客户端的数据转发交易。

首先，通过查看日志找出你的币安智能链地址地址。

确定MOP币安智能链地址后，使用BNB和MOP为您的节点提供资金。

如果时间过长，此时您可能需要重新启动节点。
### 4. 等待部署支票簿、同步voucher数据

首次启动时，客户端必须将支票簿部署到币安智能区块链(Binance Smart Chain)，并同步voucher数据，以便它可以在存储或转发时检查块的有效性。

这可能需要一段时间，所以请耐心等待！完成后，您将看到开始添加对等点并连接到网络。

您可以在这发生时通过查看日志来关注进度。

### 5. 检查客户端正在运行
MOP客户端获得资金、部署支票簿、同步存储后, HTTP API开始监听端口1683。
可通过向主机端口1683发送GET请求，检查一切是否按预期工作。
```shell script
curl localhost:1633
```

