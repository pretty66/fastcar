![logo](./docs/fastcar-logo.png)

[中文](./README.md) &nbsp;&nbsp;&nbsp;&nbsp; [English](./README_EN.md)
----

# Fastcar - PHP Long Connection Proxy

Fastcar is a proxy program written in Go language, designed to provide long connection support for PHP programs when making network requests. Through Unix domain socket communication, Fastcar can empower PHP programs so that they can maintain long connections, thereby improving network request performance.

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
## Introduction and technical principles

Fastcar is deployed on the server where the PHP program is located, and communicates with the PHP program through Unix domain sockets to achieve efficient connections. Its core competencies include:

- **Communication Method**: Efficient communication is achieved using Unix domain sockets, reducing network overhead.

- **Determining the Request Protocol**: Fastcar receives network requests from PHP programs via Unix domain sockets and determines whether it is an HTTP or HTTPS request based on the request protocol. For HTTP requests, Fastcar forwards them directly to remote servers. For HTTPS requests, Fastcar temporarily impersonates the server, completes the TLS handshake with the PHP program through in-memory communication, and then establishes a TLS connection with the remote server to facilitate request forwarding.

- **Connection Pool**: Fastcar maintains a connection pool to recycle connections to remote servers, enabling connection reuse and reducing resource consumption.

## Challenges in PHP Network Requests
In traditional network requests, PHP programs use a short connection mode, which can lead to the following issues:

- **Performance Problems**: Frequent connection establishment and closure can lead to performance degradation, affecting the overall response speed of PHP programs.
- **Increased Latency**: Short connection mode can increase request latency, especially in high-concurrency scenarios, where connection overhead becomes significant.
- **Resource Waste**: Frequent connection establishment and closure waste server and network resources, affecting the efficient utilization of resources.

## Advantages of Fastcar

Fastcar offers multiple advantages that make it an ideal choice for PHP long-connection proxy:

- **HTTPS Handshake Impersonation**: When accessing HTTPS services, Fastcar impersonates the server for TLS handshake, replaces the PHP program in establishing the TLS connection with the target service, and maintains connection reuse. This significantly reduces request latency caused by frequent handshake in short PHP connections.

- **Long Connection Support**: By maintaining a connection pool, Fastcar provides support for long connections. It can handle network requests for all PHP programs on a node, achieving multiple-service connections sharing. For example, when deploying multiple pods on a single node in a Kubernetes-like setup, it greatly reduces connection establishment overhead.

- **Resource Optimization**: Fastcar itself is very lightweight and consumes minimal system resources. Additionally, through connection pool reuse, it reduces the consumption of server and network resources.

## Performance Comparison
Requests to the Baidu Translate API were made to both HTTP and HTTPS endpoints, and the delay comparisons between using Fastcar and making direct PHP requests are shown in the following graphs:

### HTTP Request Delay Comparison
100 requests were made directly and through Fastcar. With Fastcar, the average delay decreased by approximately 34.8%.
![HTTP Delay](./docs/http_delay.png)

### HTTPS Request Delay Comparison
100 requests were made directly and through Fastcar. With Fastcar, the average delay decreased by approximately 48.4%.
![HTTPS Delay](./docs/https_delay.png)


### Testing has shown that Fastcar can significantly reduce network request latency for PHP programs, making the delay more stable. This advantage is even more pronounced in HTTPS request mode.





## Quick start
### Deploy fastcar
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

### PHP program configuration
```php
// Disable certificate validation
curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);
// Set Unix socket file path
curl_setopt($ch, CURLOPT_UNIX_SOCKET_PATH, '/tmp/fastcar.sock');
```
Reference example：[Access Baidu Translation API](./example/curl.php)

## Contribute

Feel free to ask questions, report bugs, or suggest improvements to Fastcar. You can communicate with us through GitHub Issues.

## License
Fastcar is under the Apache 2.0 license. See the [LICENSE](./LICENSE) directory for details.
