# 源代码构建

MOP是使用[Go](https://golang.org/)语言编写的。您可以直接从[源代码](https://github.com/redesblock/mop)构建客户端软件。
> 从源安装的先决条件：
> - go - 从[golang.org](https://golang.org/dl)下载最新版本。
> - git - 从[git-scm.com](https://git-scm.com/)下载。
> - make - 通常包含在大多数操作系统中。

## 直接下载
官方源码编译的程序[releases](https://github.com/redesblock/mop/releases), 您可以根据操作系统选择相应的下载。

## 源代码构建

1. 克隆存储库
   ```shell script
   git clone https://github.com/redesblock/mop.git
   cd mop
   ```
2. 查找最新版本
   ```shell script
   git describe --tags
   ```
3. 切换所需的版本
   ```shell script
   git checkout v0.9.1
   ```
4. 构建二进制文件
   ```shell script
   make binary
   ```
5. 验证二进制文件
   ```shell script
   bin/mop version
   ```
6. 将二进制文件移动到您$PATH目录
   ```shell script
    sudo cp bin/mop /usr/local/bin/mop 
   ```