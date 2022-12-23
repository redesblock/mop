# 接口文档

MOP节点提供了`API`和`Debug API`两个 HTTP API 服务。可以使用熟悉的HTTP请求调用相应接口，并以语义准确的HTTP 状态和错误代码以及JSON格式的数据有效负载进行响应。

## API

API服务提供了向 MOP网络上传和下载内容的所有功能。默认情况下，它在端口上运行`:1683`。更多详细信息[API](https://redesblock.github.io/mop/api.html)

## Debug API

Debug API 提供了节点运行时检查其状态的功能，以及一些不应公开给公共互网络的其他功能。默认情况下，Debug API 在端口上运行`:1685`。默认情况下，Debug API 是禁用的，但可以通过将`debug-api-enable`设置为`true`启用。更多详细信息[Debug API](https://redesblock.github.io/mop/debug-api.html)

> Debug API 不应暴露在公共互联网上，确保您的网络有阻止端口的防火墙`1685`，或将调试 API 绑定到`localhost`