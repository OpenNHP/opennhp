package etcd

import (
	"context"
	"errors"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/OpenNHP/opennhp/nhp/log"
)

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
	watcher   clientv3.Watcher
	signals   struct {
		stop chan struct{}
	}
}

func (conn *EtcdConn) InitClient() error {
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn.ctx = ctx
	_, err = conn.client.Status(conn.ctx, conn.Endpoints[0])
	if err != nil {
		return err
	}
	return nil
}

func (conn *EtcdConn) GetValue() ([]byte, error) {
	val, err := conn.client.Get(conn.ctx, conn.Key)
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
	_, err := conn.client.Put(conn.ctx, conn.Key, v)
	return err
}

func (conn *EtcdConn) WatchValue(callbackFunc func(val []byte)) {
	// create etcd watcher
	conn.watcher = clientv3.NewWatcher(conn.client)

	watchChan := conn.watcher.Watch(context.Background(), conn.Key)

	for {
		select {
		case resp := <-watchChan:
			// handle change events
			for _, ev := range resp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					callbackFunc(ev.Kv.Value)
				case clientv3.EventTypeDelete:
					log.Debug("[DELETE] Key: %s\n", string(ev.Kv.Key))
				}
			}
		case <-conn.signals.stop:
			conn.watcher.Close()
			return
		}
	}

}

func (conn *EtcdConn) Close() {
	if conn.client != nil {
		// stop the etcd watcher
		close(conn.signals.stop)
		conn.client.Close()
	}
}
