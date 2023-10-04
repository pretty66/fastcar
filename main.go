package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

// BuildDate: Binary file compilation time
var (
	BuildDate string
)

var (
	tlsc      *tls.Config
	transport *http.Transport
	revers    *httputil.ReverseProxy
)

type Options struct {
	UnixSocket string
	ProxyURL   string
	CertFile   string
	KeyFile    string
}

var options *Options

func init() {
	options = &Options{}
	flag.StringVar(&options.UnixSocket, "unix", "/tmp/fastcar.sock", "unix socket addr")
	flag.StringVar(&options.ProxyURL, "proxy", "", "proxy url: scheme://user:password@host:port")
	flag.StringVar(&options.CertFile, "cert", "", "tls cert file, empty auto create")
	flag.StringVar(&options.KeyFile, "key", "", "tls key file, empty auto create")
	flag.Parse()

	// ========== transport ============
	transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          1024,
		IdleConnTimeout:       120 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   100,
	}

	if options.ProxyURL != "" {
		pu, err := url.Parse(options.ProxyURL)
		if err != nil {
			panic(fmt.Errorf("invalid proxy url: %w", err))
		}
		transport.Proxy = http.ProxyURL(pu)
		transport.ForceAttemptHTTP2 = false
	}
	// ========== revers proxy ============
	revers = &httputil.ReverseProxy{
		FlushInterval: -1,
		Director: func(req *http.Request) {
			req.URL.Host = req.Host
			localAddr := req.Context().Value(http.LocalAddrContextKey).(net.Addr).String()
			index := strings.Index(localAddr, ":")
			if index < 4 {
				log.Println("BUG: Local address format error:", localAddr)
				return
			}
			req.URL.Scheme = localAddr[:index]
		},
		Transport: transport,
	}

	// ========== tls config ============
	var cf tls.Certificate
	var err error
	if options.CertFile != "" && options.KeyFile != "" {
		cf, err = tls.LoadX509KeyPair(options.CertFile, options.KeyFile)
		if err != nil {
			panic(fmt.Errorf("load cert file error: %w", err))
		}
	} else {
		key, cert, err := createCert()
		if err != nil {
			panic(fmt.Errorf("create cert error: %w", err))
		}
		cf = tls.Certificate{PrivateKey: key, Certificate: [][]byte{cert.Raw}}
	}

	tlsc = &tls.Config{
		Certificates: []tls.Certificate{cf},
	}
}

func main() {
	_ = os.RemoveAll(options.UnixSocket)
	ln, err := net.Listen("unix", options.UnixSocket)
	if err != nil {
		panic(err)
	}
	fl := &FastCarListener{ln}
	showBanner()
	err = http.Serve(fl, revers)
	if err != nil {
		panic(err)
	}
}

func showBanner() {
	fmt.Println("running fastcar, listen on unix socket:", options.UnixSocket)
	fmt.Println("Build Date: ", BuildDate)
	fmt.Println(` ________         ______  
|        \       /      \ 
| $$$$$$$$      |  $$$$$$\
| $$__          | $$   \$$
| $$  \         | $$      
| $$$$$         | $$   __ 
| $$            | $$__/  \
| $$             \$$    $$
 \$$              \$$$$$$ 
                          
                          
                          `)
}

type bufConn struct {
	net.Conn
	buf    *bytes.Reader
	scheme string
}

func (c *bufConn) Read(p []byte) (int, error) {
	if c.buf == nil {
		return c.Conn.Read(p)
	}
	n, err := c.buf.Read(p)
	if err != nil {
		return n, err
	}
	if n < len(p) {
		c.buf = nil
		cn, err := c.Conn.Read(p[n:])
		cn += n
		return cn, err
	}

	return n, nil
}

func (c *bufConn) LocalAddr() net.Addr {
	name := c.scheme + ":" + options.UnixSocket
	return &net.UnixAddr{Name: name, Net: "unix"}
}

type FastCarListener struct {
	net.Listener
}

func (fl *FastCarListener) Accept() (net.Conn, error) {
	conn, err := fl.Listener.Accept()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 1)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		log.Println("first read err: ", err)
		return nil, err
	}
	bconn := &bufConn{Conn: conn, buf: bytes.NewReader(buf)}
	switch buf[0] {
	case 0x16:
		tlsConn := tls.Server(bconn, tlsc)
		if err := tlsConn.Handshake(); err != nil {
			return nil, err
		}
		conn = tlsConn
		bconn.scheme = "https"
	default:
		conn = bconn
		bconn.scheme = "http"
	}
	return conn, nil
}

func createCert() (*ecdsa.PrivateKey, *x509.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano() / 100000),
		Subject: pkix.Name{
			CommonName:   "fastcar",
			Organization: []string{"fastcar"},
		},
		NotBefore:             time.Now().Add(-time.Hour * 48),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365 * 100),
		BasicConstraintsValid: true,
		IsCA:                  false,
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, err
	}

	return key, cert, nil
}
