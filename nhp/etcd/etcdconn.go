package etcd

import (
	"context"
	"errors"
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
}

type EtcdConn struct {
	Endpoints []string
	Username  string
	Password  string
	Key       string
	client    *clientv3.Client
	ctx       context.Context
	cancel    context.CancelFunc
}

func (conn *EtcdConn) InitClient() error {
	if len(conn.Endpoints) == 0 {
		return errors.New("no etcd endpoints configured")
	}
	var err error
	conn.client, err = clientv3.New(clientv3.Config{
		Endpoints:   conn.Endpoints,
		DialTimeout: 5 * time.Second,
		Username:    conn.Username,
		Password:    conn.Password,
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
					log.Error("etcd watch for %q ended unexpectedly: %v; remote configuration updates have stopped", conn.Key, resp.Err())
				}
				return
			}
			// handle change events
			for _, ev := range resp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					callbackFunc(ev.Kv.Value)
				case clientv3.EventTypeDelete:
					log.Debug("[DELETE] Key: %s\n", string(ev.Kv.Key))
				}
			}
		case <-conn.ctx.Done():
			return
		}
	}

}

func (conn *EtcdConn) Close() {
	if conn.cancel != nil {
		conn.cancel()
	}
	if conn.client != nil {
		conn.client.Close()
	}
}
