// Package startuptest supplies a loopback etcd protocol fixture for daemon startup tests.
package startuptest

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Remote struct {
	etcdserverpb.UnimplementedKVServer
	etcdserverpb.UnimplementedMaintenanceServer
	etcdserverpb.UnimplementedWatchServer
	Value        string
	SecondValue  string
	FailSecond   bool
	FailStatus   bool
	Updates      chan []byte
	WatchStarted chan struct{}
	reads        atomic.Int32
}

func (r *Remote) Status(context.Context, *etcdserverpb.StatusRequest) (*etcdserverpb.StatusResponse, error) {
	if r.FailStatus {
		return nil, status.Error(codes.PermissionDenied, "startup status denied")
	}
	return &etcdserverpb.StatusResponse{Header: &etcdserverpb.ResponseHeader{}}, nil
}

func (r *Remote) Range(context.Context, *etcdserverpb.RangeRequest) (*etcdserverpb.RangeResponse, error) {
	n := r.reads.Add(1)
	if n == 2 && r.FailSecond {
		return nil, status.Error(codes.FailedPrecondition, "initial remote read refused")
	}
	value := r.Value
	if n == 2 && r.SecondValue != "" {
		value = r.SecondValue
	}
	return &etcdserverpb.RangeResponse{Header: &etcdserverpb.ResponseHeader{Revision: 1}, Kvs: []*mvccpb.KeyValue{{Key: []byte("/startup"), Value: []byte(value)}}}, nil
}

func (r *Remote) Watch(stream etcdserverpb.Watch_WatchServer) error {
	if _, err := stream.Recv(); err != nil {
		return err
	}
	if err := stream.Send(&etcdserverpb.WatchResponse{Created: true, Header: &etcdserverpb.ResponseHeader{Revision: 1}}); err != nil {
		return err
	}
	if r.WatchStarted != nil {
		close(r.WatchStarted)
	}
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case value := <-r.Updates:
			if err := stream.Send(&etcdserverpb.WatchResponse{Header: &etcdserverpb.ResponseHeader{Revision: 2}, Events: []*mvccpb.Event{{Type: mvccpb.PUT, Kv: &mvccpb.KeyValue{Key: []byte("/startup"), Value: value}}}}); err != nil {
				return err
			}
		}
	}
}

func (r *Remote) Serve(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	etcdserverpb.RegisterKVServer(s, r)
	etcdserverpb.RegisterMaintenanceServer(s, r)
	etcdserverpb.RegisterWatchServer(s, r)
	go func() { _ = s.Serve(l) }()
	t.Cleanup(s.Stop)
	return "http://" + l.Addr().String()
}

func Directory(t *testing.T, remote string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "etc"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "etc", "remote.toml"), []byte(remote), 0600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func Config(endpoint string) string {
	return fmt.Sprintf("Provider = 'etcd'\nEndpoints = [%q]\nKey = 'startup'\n", endpoint)
}

// Startup unit tests must never execute the host's firewall utilities.
func FakeFirewall(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"iptables", "ip6tables", "ipset"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\nexit 0\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(filepath.Join(dir, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}
