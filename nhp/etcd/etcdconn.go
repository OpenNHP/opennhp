package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/OpenNHP/opennhp/nhp/log"
)

const requestTimeout = 3 * time.Second

type EtcdConfig struct {
	Key       string
	Endpoints []string
	Username  string
	Password  string

	// TLS for the connection to etcd. Optional, and off when unset: a
	// deployment that says nothing keeps the plaintext connection it had.
	//
	// It is worth turning on. The value under Key is the AC's and the
	// server's trusted-peer configuration -- an attacker on the path can
	// rewrite it to install their own nhp-server public key as trusted,
	// which is a full bypass of the access control this product exists to
	// provide. Username and Password are also sent in the clear without
	// it, so they authenticate the client to anyone listening as much as
	// to etcd.
	//
	// CAFile alone gives server authentication and confidentiality.
	// CertFile and KeyFile add client certificate authentication, and must
	// be given together.
	CAFile   string
	CertFile string
	KeyFile  string
}

type EtcdConn struct {
	Endpoints  []string
	Username   string
	Password   string
	Key        string
	CAFile     string
	CertFile   string
	KeyFile    string
	client     *clientv3.Client
	ctx        context.Context
	cancel     context.CancelFunc
	callbackMu sync.Mutex
}

// InitClient must be called once per EtcdConn, before concurrent use.
// After an initialization failure or Close, create a new EtcdConn to retry.
func (conn *EtcdConn) InitClient() error {
	if len(conn.Endpoints) == 0 {
		return errors.New("no etcd endpoints configured")
	}
	tlsConf, err := conn.tlsConfig()
	if err != nil {
		return err
	}
	if tlsConf == nil {
		// Said once, at startup, where an operator reading the log can act
		// on it. The remote value is the trusted-peer configuration; see
		// EtcdConfig for what rewriting it buys an attacker on the path.
		log.Warning("etcd connection to %v is not using TLS: the remote configuration, and the username and password, travel in the clear. Set CAFile in remote.toml to enable TLS.", conn.Endpoints)
	}

	conn.client, err = clientv3.New(clientv3.Config{
		Endpoints:   conn.Endpoints,
		DialTimeout: 5 * time.Second,
		Username:    conn.Username,
		Password:    conn.Password,
		TLS:         tlsConf,
	})

	conn.Key = "/" + conn.Key
	if err != nil {
		return err
	}
	conn.ctx, conn.cancel = context.WithCancel(context.Background())
	ctx, cancel := context.WithTimeout(conn.ctx, requestTimeout)
	defer cancel()
	_, err = conn.client.Status(ctx, conn.Endpoints[0])
	if err != nil {
		conn.Close()
		return err
	}
	return nil
}

func (conn *EtcdConn) GetValue() ([]byte, error) {
	if conn.client == nil || conn.ctx == nil {
		return nil, errors.New("etcd client not initialized")
	}
	ctx, cancel := context.WithTimeout(conn.ctx, requestTimeout)
	defer cancel()
	val, err := conn.client.Get(ctx, conn.Key)
	if err != nil {
		return nil, err
	}

	if len(val.Kvs) == 0 {
		return nil, errors.New("key not found")
	}
	if len(val.Kvs[0].Value) == 0 {
		return nil, errors.New("value not set")
	}
	return val.Kvs[0].Value, nil
}

func (conn *EtcdConn) SetValue(v string) error {
	if conn.client == nil || conn.ctx == nil {
		return errors.New("etcd client not initialized")
	}
	ctx, cancel := context.WithTimeout(conn.ctx, requestTimeout)
	defer cancel()
	_, err := conn.client.Put(ctx, conn.Key, v)
	return err
}

// WatchValue delivers callbacks serially. A callback must return before Close
// can finish, and must not call Close synchronously.
func (conn *EtcdConn) WatchValue(callbackFunc func(val []byte)) {
	if conn.client == nil || conn.ctx == nil {
		log.Error("cannot watch etcd configuration: client not initialized")
		return
	}
	// create etcd watcher
	watcher := clientv3.NewWatcher(conn.client)
	defer watcher.Close()
	watchChan := watcher.Watch(conn.ctx, conn.Key)

	for {
		select {
		case resp, ok := <-watchChan:
			if !ok || resp.Canceled {
				if conn.ctx.Err() == nil {
					err := resp.Err()
					if !ok {
						err = errors.New("watch channel closed")
					}
					log.Error("etcd watch for %q ended unexpectedly: %v; remote configuration updates have stopped", conn.Key, err)
				}
				return
			}
			// handle change events
			for _, ev := range resp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					func() {
						conn.callbackMu.Lock()
						defer conn.callbackMu.Unlock()
						if conn.ctx.Err() == nil {
							callbackFunc(ev.Kv.Value)
						}
					}()
				case clientv3.EventTypeDelete:
					log.Debug("[DELETE] Key: %s\n", string(ev.Kv.Key))
				}
			}
		case <-conn.ctx.Done():
			return
		}
	}

}

// Close cancels requests and waits for any active configuration callback.
func (conn *EtcdConn) Close() {
	if conn.cancel != nil {
		conn.cancel()
	}
	conn.callbackMu.Lock()
	defer conn.callbackMu.Unlock()
	if conn.client != nil {
		conn.client.Close()
	}
}

// tlsConfig builds the TLS configuration for the etcd client, or nil when
// the deployment did not ask for TLS.
//
// Not forced: a deployment that configures nothing keeps the plaintext
// connection it has always had. What is refused is a half-configured one,
// because the failure mode of guessing is a connection that looks
// encrypted and is not, or one that silently accepts any server.
func (conn *EtcdConn) tlsConfig() (*tls.Config, error) {
	if conn.CAFile == "" && conn.CertFile == "" && conn.KeyFile == "" {
		return nil, nil
	}

	// A client certificate is a pair. One half alone is a configuration
	// mistake, not a weaker setting.
	if (conn.CertFile == "") != (conn.KeyFile == "") {
		return nil, errors.New("etcd TLS: CertFile and KeyFile must be set together")
	}
	// Without a CA the client would fall back to the system roots, which
	// for an internal etcd almost never contain its issuer: refuse rather
	// than present a client certificate to a server nothing vouches for.
	if conn.CAFile == "" {
		return nil, errors.New("etcd TLS: CAFile is required when CertFile and KeyFile are set")
	}

	caPEM, err := os.ReadFile(conn.CAFile)
	if err != nil {
		return nil, fmt.Errorf("etcd TLS: cannot read CAFile %q: %w", conn.CAFile, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("etcd TLS: CAFile %q contains no usable certificate", conn.CAFile)
	}

	conf := &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
	}

	if conn.CertFile != "" {
		cert, err := tls.LoadX509KeyPair(conn.CertFile, conn.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("etcd TLS: cannot load the client certificate: %w", err)
		}
		conf.Certificates = []tls.Certificate{cert}
	}

	return conf, nil
}
