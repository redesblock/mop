# 脚本安装

我们提供了一个方便的安装脚本，它会自动检测您的执行环境并在您的计算机上安装最新稳定版本的客户端。

如果您的系统不受支持，您可能想尝试直接从[源代码构建](installation/install-from-source.md)。

## 脚本安装

1. 要使用脚本安装二进制文件，请在终端中运行以下命令之一：
    - curl
    ```shell
    curl -s https://raw.githubusercontent.com/redesblock/mop/main/install.sh | TAG=v0.9.2 bash
    ```
    - wget
    ```shell
    wget -q -O - https://raw.githubusercontent.com/redesblock/mop/main/install.sh | TAG=v0.9.2 bash
    ```
2. 脚本执行完成后，可以运行测试是否安装成功。
    ```shell script
    mop version
    ```