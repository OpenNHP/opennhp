package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

var testConnInitTime atomic.Int64

func newDBConnTestServer(peers ...*core.UdpPeer) *UdpServer {
	s := &UdpServer{
		dbConnectionMap: make(map[string]*DBConn),
		dbPeerMap:       make(map[string]*core.UdpPeer),
	}
	for _, p := range peers {
		s.dbPeerMap[p.PubKeyBase64] = p
	}
	return s
}

func testDBPeer(key string) *core.UdpPeer {
	return &core.UdpPeer{PubKeyBase64: key, Type: core.NHP_DB}
}

func testDBConn(peer *core.UdpPeer, dbId string) *DBConn {
	return &DBConn{
		ConnData: &core.ConnectionData{InitTime: testConnInitTime.Add(1)},
		DBPeer:   peer,
		DBId:     dbId,
	}
}

// A dbId belongs to the nhp-db that came online with it: another nhp-db
// cannot take its route over while it stays configured and connected.
func TestBindDBConnRefusesAnotherHolder(t *testing.T) {
	a, b := testDBPeer("key-a"), testDBPeer("key-b")
	s := newDBConnTestServer(a, b)

	connA := testDBConn(a, "db-1")
	if err := s.bindDBConn(connA); err != nil {
		t.Fatal(err)
	}
	if err := s.bindDBConn(testDBConn(b, "db-1")); !errors.Is(err, errDbIdHeld) {
		t.Fatalf("takeover by another nhp-db: got %v, want errDbIdHeld", err)
	}
	if c, err := s.dbConnForObject("db-1", ""); err != nil || c != connA {
		t.Fatal("route changed after a refused takeover")
	}
	if got := s.dbIdHolder("db-1"); got != "key-a" {
		t.Fatalf("holder = %q, want key-a", got)
	}
}

// The same nhp-db coming back over a new connection replaces its own route,
// and teardown of the old connection leaves the new route alone.
func TestBindDBConnSameHolderReconnects(t *testing.T) {
	a := testDBPeer("key-a")
	s := newDBConnTestServer(a)

	oldConn, newConn := testDBConn(a, "db-1"), testDBConn(a, "db-1")
	if err := s.bindDBConn(oldConn); err != nil {
		t.Fatal(err)
	}
	if err := s.bindDBConn(newConn); err != nil {
		t.Fatalf("reconnect: %v", err)
	}
	s.unbindDBConn(oldConn.ConnData)
	if c, err := s.dbConnForObject("db-1", "key-a"); err != nil || c != newConn {
		t.Fatal("old connection's teardown removed the new route")
	}
	s.unbindDBConn(newConn.ConnData)
	if _, err := s.dbConnForObject("db-1", "key-a"); err == nil {
		t.Fatal("route survived its connection's teardown")
	}
}

// An nhp-db holds one dbId; coming online with another releases the first.
func TestBindDBConnChangesDbId(t *testing.T) {
	a, b := testDBPeer("key-a"), testDBPeer("key-b")
	s := newDBConnTestServer(a, b)

	if err := s.bindDBConn(testDBConn(a, "db-1")); err != nil {
		t.Fatal(err)
	}
	if err := s.bindDBConn(testDBConn(a, "db-2")); err != nil {
		t.Fatal(err)
	}
	if got := s.dbIdHolder("db-1"); got != "" {
		t.Fatalf("db-1 still held by %q", got)
	}
	if err := s.bindDBConn(testDBConn(b, "db-1")); err != nil {
		t.Fatalf("released dbId: %v", err)
	}
}

// A holder that was removed from db.toml, or has expired, no longer holds
// its dbId: it gets no DWR, and a configured nhp-db can take the dbId.
func TestDBConnHolderMustStayActive(t *testing.T) {
	for name, retire := range map[string]func(s *UdpServer, p *core.UdpPeer){
		"removed": func(s *UdpServer, p *core.UdpPeer) { delete(s.dbPeerMap, p.PubKeyBase64) },
		"expired": func(s *UdpServer, p *core.UdpPeer) { p.ExpireTime = time.Now().Add(-time.Hour).Unix() },
	} {
		t.Run(name, func(t *testing.T) {
			a, b := testDBPeer("key-a"), testDBPeer("key-b")
			s := newDBConnTestServer(a, b)
			if err := s.bindDBConn(testDBConn(a, "db-1")); err != nil {
				t.Fatal(err)
			}
			retire(s, a)

			if _, err := s.dbConnForObject("db-1", ""); !errors.Is(err, errDbIdHolderInactive) {
				t.Fatalf("DWR to an inactive nhp-db: got %v, want errDbIdHolderInactive", err)
			}
			if got := s.dbIdHolder("db-1"); got != "" {
				t.Fatalf("inactive nhp-db still holds the dbId: %q", got)
			}
			if err := s.bindDBConn(testDBConn(b, "db-1")); err != nil {
				t.Fatalf("takeover from an inactive nhp-db: %v", err)
			}
		})
	}
}

// An owned object's DWR only goes to its owner, whoever holds the dbId.
func TestDBConnForObjectChecksOwner(t *testing.T) {
	a := testDBPeer("key-a")
	s := newDBConnTestServer(a)
	if err := s.bindDBConn(testDBConn(a, "db-1")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.dbConnForObject("db-1", "key-b"); !errors.Is(err, errDbIdHolderNotOwner) {
		t.Fatalf("DWR for key-b's object held by key-a: got %v, want errDbIdHolderNotOwner", err)
	}
	if _, err := s.dbConnForObject("db-1", "key-a"); err != nil {
		t.Fatal("owner's object not routed")
	}
	if _, err := s.dbConnForObject("db-1", ""); err != nil {
		t.Fatal("legacy object not routed by dbId")
	}
	if _, err := s.dbConnForObject("db-2", ""); !errors.Is(err, errDbIdNotOnline) {
		t.Fatalf("unknown dbId: got %v, want errDbIdNotOnline", err)
	}
}

// Regression for the dbConnectionMap race: DAV lookups, DOL binds and
// connection teardown on separate goroutines. Run with -race; before the
// fix an unlocked read here was a "concurrent map read and map write"
// fatal error that took nhp-server down.
func TestDBConnConcurrentAccess(t *testing.T) {
	const peers, rounds = 8, 200
	var ps []*core.UdpPeer
	for i := 0; i < peers; i++ {
		ps = append(ps, testDBPeer(fmt.Sprintf("key-%d", i)))
	}
	s := newDBConnTestServer(ps...)

	var wg sync.WaitGroup
	for i, p := range ps {
		wg.Add(2)
		go func(p *core.UdpPeer, dbId string) {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				c := testDBConn(p, dbId)
				_ = s.bindDBConn(c)
				s.unbindDBConn(c.ConnData)
			}
		}(p, fmt.Sprintf("db-%d", i%3))
		go func(dbId, owner string) {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				s.dbConnForObject(dbId, owner)
				s.dbIdHolder(dbId)
			}
		}(fmt.Sprintf("db-%d", i%3), p.PubKeyBase64)
	}
	wg.Wait()
}

// dbOnlineTestServer is a server with nhp-db peers keyed by the base64 of
// the raw keys, so HandleDBOnline finds them from ppd.RemotePubKey.
func dbOnlineTestServer(t *testing.T, rawKeys ...string) (*UdpServer, map[string]*core.UdpPeer) {
	t.Helper()
	s := newPerIPCapTestServer(t)
	s.dbConnectionMap = make(map[string]*DBConn)
	s.dbPeerMap = make(map[string]*core.UdpPeer)
	peers := make(map[string]*core.UdpPeer)
	for _, raw := range rawKeys {
		p := testDBPeer(base64.StdEncoding.EncodeToString([]byte(raw)))
		s.dbPeerMap[p.PubKeyBase64] = p
		peers[raw] = p
	}
	return s, peers
}

// dbOnline delivers an NHP_DOL for dbId from rawKey over a newly admitted
// direct connection, and returns the connection and HandleDBOnline's error.
func dbOnline(t *testing.T, s *UdpServer, rawKey, dbId string, port int) (*UdpConn, error) {
	t.Helper()
	conn := newPerIPCapTestConn("203.0.113.40", port)
	conn.ConnData.InitTime = testConnInitTime.Add(1)
	if reason := s.admitDirectConnection(conn, conn.ConnData.RemoteAddr.String()); reason != "" {
		t.Fatalf("admission refused: %s", reason)
	}
	body, err := json.Marshal(&common.DBOnlineMsg{DBId: dbId})
	if err != nil {
		t.Fatal(err)
	}
	ppd := &core.PacketParserData{
		ConnData:     conn.ConnData,
		BodyMessage:  body,
		RemotePubKey: []byte(rawKey),
		SenderTrxId:  1,
	}
	return conn, s.HandleDBOnline(ppd)
}

// assertNotPromoted checks that a refused NHP_DOL left its connection an
// ordinary one: agent timeout, per-IP bucket, rate limit, no route.
func assertNotPromoted(t *testing.T, s *UdpServer, conn *UdpConn) {
	t.Helper()
	if conn.isDBConnection.Load() {
		t.Error("refused connection was promoted to a DB connection")
	}
	if conn.timeoutMs.Load() != 0 {
		t.Errorf("refused connection got timeout %d ms", conn.timeoutMs.Load())
	}
	if !conn.perIPCounted {
		t.Error("refused connection left the per-IP bucket")
	}
	if conn.rateExempt.Load() {
		t.Error("refused connection was rate-exempted")
	}
	s.dbConnectionMapMutex.Lock()
	defer s.dbConnectionMapMutex.Unlock()
	for dbId, c := range s.dbConnectionMap {
		if c.ConnData == conn.ConnData {
			t.Errorf("refused connection routes dbId %q", dbId)
		}
	}
}

// An accepted NHP_DOL binds the route, promotes the connection and exempts
// it from rate limiting. The test has no transaction to answer on, so the
// handler ends at ErrTransactionIdNotFound after all of that.
func TestHandleDBOnlineAccepts(t *testing.T) {
	s, _ := dbOnlineTestServer(t, "key-a")
	conn, err := dbOnline(t, s, "key-a", "db-1", 41000)
	if !errors.Is(err, common.ErrTransactionIdNotFound) {
		t.Fatalf("HandleDBOnline = %v, want ErrTransactionIdNotFound (no transaction)", err)
	}
	if c, err := s.dbConnForObject("db-1", ""); err != nil || c.ConnData != conn.ConnData {
		t.Fatal("accepted nhp-db is not the route for its dbId")
	}
	if !conn.isDBConnection.Load() || conn.timeout() != DefaultDBConnectionTimeoutMs {
		t.Fatal("accepted connection was not promoted")
	}
	if !conn.rateExempt.Load() {
		t.Fatal("accepted connection was not rate-exempted")
	}
}

// Every refusal happens before promotion and rate exemption, and leaves the
// existing route alone.
func TestHandleDBOnlineRefusesBeforePromotion(t *testing.T) {
	t.Run("empty dbId", func(t *testing.T) {
		s, _ := dbOnlineTestServer(t, "key-a")
		conn, err := dbOnline(t, s, "key-a", "", 41000)
		if !errors.Is(err, errEmptyDBId) {
			t.Fatalf("empty dbId: got %v, want errEmptyDBId", err)
		}
		assertNotPromoted(t, s, conn)
	})
	t.Run("whitespace around dbId", func(t *testing.T) {
		s, _ := dbOnlineTestServer(t, "key-a")
		conn, err := dbOnline(t, s, "key-a", " db-1", 41000)
		if !errors.Is(err, errDBIdWhitespace) {
			t.Fatalf("whitespace dbId: got %v, want errDBIdWhitespace", err)
		}
		assertNotPromoted(t, s, conn)
	})

	t.Run("unknown peer", func(t *testing.T) {
		s, _ := dbOnlineTestServer(t, "key-a")
		conn, err := dbOnline(t, s, "stranger", "db-1", 41001)
		if err == nil {
			t.Fatal("unknown peer accepted")
		}
		assertNotPromoted(t, s, conn)
	})

	t.Run("expired peer", func(t *testing.T) {
		s, peers := dbOnlineTestServer(t, "key-a")
		peers["key-a"].ExpireTime = time.Now().Add(-time.Hour).Unix()
		conn, err := dbOnline(t, s, "key-a", "db-1", 41002)
		if !errors.Is(err, errNotDBPeer) {
			t.Fatalf("expired peer: got %v, want errNotDBPeer", err)
		}
		assertNotPromoted(t, s, conn)
	})

	t.Run("dbId held by another nhp-db", func(t *testing.T) {
		s, _ := dbOnlineTestServer(t, "key-a", "key-b")
		holder, err := dbOnline(t, s, "key-a", "db-1", 41003)
		if !errors.Is(err, common.ErrTransactionIdNotFound) {
			t.Fatalf("holder online: %v", err)
		}
		conn, err := dbOnline(t, s, "key-b", "db-1", 41004)
		if !errors.Is(err, errDbIdHeld) {
			t.Fatalf("takeover: got %v, want errDbIdHeld", err)
		}
		assertNotPromoted(t, s, conn)
		if c, err := s.dbConnForObject("db-1", ""); err != nil || c.ConnData != holder.ConnData {
			t.Fatal("refused takeover changed the route")
		}
	})
}
