![logo](./docs/fastcar-logo.png)

[中文](./README_ZH.md) &nbsp;&nbsp;&nbsp;&nbsp; [English](./README.md)
---

# Fastcar - PHP 长连接代理

Fastcar 是使用 Go 语言编写的代理程序，旨在为 PHP 程序在网络请求时提供长连接支持。通过 Unix 域套接字通信，Fastcar 可以为 PHP 程序赋能，让它们能够保持长连接，从而提升网络请求性能。
```text
+---------------------+              +----------------------+               +---------------------+
|                     |              |                      |               |                     |
|    PHP Program      |   <------->  |       FastCar        |    <------>   |   Remote Server     |
|                     |    unix      |                      |    HTTP/HTTPS |                     |
|    Network Request  |    domain    |   +--------------+   |    Request    |   Network Request   |
|                     |    socket    |   |              |   |               |                     |
+---------------------+              |   | Connection   |   |               +---------------------+
                                     |   |   Pool       |   |
                                     |   +--------------+   |
                                     +----------------------+
```
## 简介与技术原理

Fastcar 部署在 PHP 程序所在的服务器上，通过 Unix 域套接字与 PHP 程序进行通信，实现高效的连接。其核心能力包括：

- **通信方式**：使用 Unix 域套接字（unix domain socket）进行高效通信，降低了网络开销。

- **判断请求协议**：Fastcar 从 Unix 域套接字接收 PHP 程序的网络请求，根据请求协议判断是 HTTP 请求还是 HTTPS 请求;
对于 HTTP 请求，Fastcar 直接转发给远程服务器; 对于 HTTPS 请求，Fastcar 会暂时伪装成服务端，与 PHP 程序在内存通信中完成 TLS 握手，然后再与远程服务器建立 TLS 连接，实现请求转发。

- **连接池**：Fastcar 维护连接池，将与远程服务器的连接回收，实现连接的复用，降低资源消耗。

## PHP 网络请求的挑战
在传统的网络请求中，PHP 程序使用短连接模式，这可能导致以下问题：
- 性能问题：频繁的连接建立和关闭导致性能下降，影响 PHP 程序的整体响应速度。
- 延迟增加：短连接模式会增加请求的延迟，尤其在高并发场景下，连接开销变得显著。
- 资源浪费：频繁的连接开销和关闭会浪费服务器和网络资源，影响资源的有效利用。

## Fastcar 的优势

Fastcar 提供了多重优势，使其成为 PHP 长连接代理的理想选择：

- **HTTPS 伪装握手**：针对https服务的访问Fastcar会伪装服务端进行 TLS 握手，代替php程序同目标服务建立tls连接，并保持连接复用，能显著降低php短连接频繁握手导致的请求延迟。

- **长连接支持**：通过维护连接池，Fastcar 提供长连接支持，可以接管节点上所有的php程序网络请求，达到多服务连接共用；例如k8s形式单台节点上部署多个pod形式的业务能大大降低连接建立的开销。

- **资源优化**：Fastcar本身非常轻量，只占用很少的系统资源，同时通过连接池的复用，降低了服务器和网络资源的消耗。

## 效果对比
分别请求百度翻译api的http和https地址，对比使用fastcar和php直接请求的延迟对比，如下图所示：

### http请求延迟对比
直接请求和通过fastcar请求各100次，使用fastcar后，平均延迟降低了约34.8%
![http延迟](./docs/http_delay.png)

### https请求延迟对比
直接请求和通过fastcar请求各100次，使用fastcar后，平均延迟降低了约48.4%
![https延迟](./docs/https_delay.png)

**测试发现，fastcar可以显著降低php程序的网络请求延迟，使请求的延迟趋于更加稳定的状态；在https请求模式下，fastcar的优势更加明显。**

## 快速开始
### 部署fastcar
```shell
$ git clone https://github.com/pretty66/fastcar.git
$ cd fastcar
$ make build
$
$ ./fastcar -h
Usage of ./fastcar:
  -cert string
        tls cert file, empty auto create
  -key string
        tls key file, empty auto create
  -proxy string
        proxy url: scheme://user:password@host:port
  -unix string
        unix socket addr (default "/tmp/fastcar.sock")
```
```shell
# run
nohup ./fastcar -unix /tmp/fastcar.sock > debug.log 2>&1 &
```

### php程序配置
```php
// Disable certificate validation
curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);
// Set Unix socket file path
curl_setopt($ch, CURLOPT_UNIX_SOCKET_PATH, '/tmp/fastcar.sock');
```
参考示例：[访问百度翻译api](./example/curl.php)

## 贡献

欢迎对 Fastcar 提出问题、报告 Bug 或者提供改进建议。你可以通过 GitHub Issues 来与我交流。

## License
Fastcar is under the Apache 2.0 license. See the [LICENSE](./LICENSE) directory for details.
