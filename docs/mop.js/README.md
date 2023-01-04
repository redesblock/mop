# MOP.JS

`mop.js`是一个允许您与本地或远程MOP节点交互的JavaScript库。



## 入门

首先，您需要安装mop.js项目

- npm

```shell
npm install @redesblock/mop.js --save
```

- yarn

```shell
yarn add @redesblock/mop.js --save
```

安装完成后，您需要导入 MOP类并创建一个连接到您的 MOP节点（这里我们假设它在本地主机上的默认端口上运行）。请注意，如果您将传递无效的 URL，构造函数将抛出异常！

```shell
import { MOP, MopDebug } from "@redesblock/mop.js"

const mop = new MOP('http://localhost:1683')
const mopDebug = new MopDebug("http://localhost:1675")
```

## 上传和下载

使用`mop.js` 很容易将您的数据上传到MOP网络。您可以直接上传非结构化数据、单个文件甚至复杂目录。让我们一一浏览这些选项。

> Voucher代券
>
> 上传数据到到MOP网络要求每个操作都需要携带Voucher代券



### 数据

您可以使用上传任何`string`或`Uint8Array`数据。

当您下载数据`downloadData`时，返回类型是`Data`接口，它扩展`Uint8Array`了方便的功能，例如：

- `text()`将字节转换为 UTF-8 编码的字符串
- `hex()`将字节转换为**无前缀的**十六进制字符串
- `json()`将字节转换为 JSON 对象

```js
const voucherBatchId = await mopDebug.createPostageBatch("100", 17)
const result = await mop.uploadData(voucherBatchId, "MOP is awesome!")
console.log(result.reference) 

const retrievedData = await mop.downloadData(result.reference)

console.log(retrievedData.text()) // prints 'MOP is awesome!'
```

### 文件

您可以上传文件并包含文件名。当您下载文件时，`mop.js`将返回附加信息，例如文件类型`contentType`或文件名称`name`

```js
const voucherBatchId = await mopDebug.createVoucherBatch("100", 17, {waitForUsable: false})
const result = await mop.uploadFile(voucherBatchId, "MOP is awesome!", "textfile.txt")
const retrievedFile = await mop.downloadFile(result.reference)

console.log(retrievedFile.name) // prints 'textfile.txt'
console.log(retrievedFile.contentType) // prints 'application/octet-stream'
console.log(retrievedFile.data.text()) // prints 'MOP is awesome!'
```

在浏览器中，您可以直接上传`File`类型。文件名取自文件对象本身

```js
const file = new File(["foo"], "foo.txt", { type: "text/plain" })

const voucherBatchId = await mopDebug.createVoucherBatch("100", 17, {waitForUsable: false})
const result = await mop.uploadFile(voucherBatchId, file)
const retrievedFile = await mop.downloadFile(result.reference)

console.log(retrievedFile.name) // prints 'foo.txt'
console.log(retrievedFile.contentType) // prints 'text/plain'
console.log(retrievedFile.data.text()) // prints 'foo'
```

## 目录

在nodejs中，您可以使用该`uploadFilesFromDirectory`功能，该功能将目录路径作为输入并上传该目录中的所有文件。假设我们有以下数据结构：

```shell
.
+-- foo.txt
+-- dir
|   +-- bar.txt
```

```js
const voucherBatchId = await mopDebug.createVoucherBatch("100", 17)

const result = await mop.uploadFilesFromDirectory(voucherBatchId, './') // upload recursively current folder

const rFoo = await mop.downloadFile(result.reference, './foo.txt') // download foo
const rBar = await mop.downloadFile(result.reference, './dir/bar.txt') // download bar

console.log(rFoo.data.text()) // prints 'foo'
console.log(rBar.data.text()) // prints 'bar'
```

在浏览器中，您可以轻松地直接从您的表单上传一组`File`列表.

```js
const foo = new File(["foo"], "foo.txt", { type: "text/plain" })
const bar = new File(["bar"], "bar.txt", { type: "text/plain" })

const voucherBatchId = await mopDebug.createVoucherBatch("100", 17)
const result = await mop.uploadFiles(voucherBatchId, [ foo, bar ]) // upload

const rFoo = await mop.downloadFile(result.reference, './foo.txt') // download foo
const rBar = await mop.downloadFile(result.reference, './bar.txt') // download bar

console.log(rFoo.data.text()) // prints 'foo'
console.log(rBar.data.text()) // prints 'bar'
```

