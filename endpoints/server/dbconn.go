package server

import (
	"errors"
	"fmt"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// errDbIdHeld refuses an NHP_DOL for a dbId that another configured nhp-db
// is online with.
var errDbIdHeld = errors.New("dbId is held by another nhp-db")

// dbConnectionMap routes NHP_DWR for a data object to the nhp-db that is
// online with the object's dbId. Every access goes through the methods in
// this file, under dbConnectionMapMutex: NHP_DAV, NHP_DOL and connection
// teardown run on different goroutines, and an unlocked read racing a write
// is a Go runtime fatal error that no recover() catches.
//
// A dbId is bound to the public key of the nhp-db that first comes online
// with it, for as long as that nhp-db stays configured, unexpired and
// connected; another nhp-db cannot take the route over. The same nhp-db
// coming back (restart, new source port) replaces its own entry.
//
// Lock order: dbConnectionMapMutex, then dbPeerMapMutex.

// activeDBPeerLocked reports whether pubKey belongs to an nhp-db that is
// listed, and not expired, in db.toml. The caller holds
// dbConnectionMapMutex; this takes dbPeerMapMutex.
func (s *UdpServer) activeDBPeerLocked(pubKey string) bool {
	s.dbPeerMapMutex.Lock()
	peer := s.dbPeerMap[pubKey]
	s.dbPeerMapMutex.Unlock()
	return peer != nil && !peer.IsExpired()
}

// dbIdHeldLocked reports whether dbId is held by an active nhp-db other
// than pubKey. The caller holds dbConnectionMapMutex.
func (s *UdpServer) dbIdHeldLocked(dbId, pubKey string) bool {
	held, found := s.dbConnectionMap[dbId]
	if !found {
		return false
	}
	heldKey := held.DBPeer.PublicKeyBase64()
	return heldKey != pubKey && s.activeDBPeerLocked(heldKey)
}

// checkDBIdAvailable is bindDBConn's refusal without the bind, for
// HandleDBOnline to run before it promotes the connection: a connection
// that will be refused should not get the longer control-connection
// timeout or leave the per-IP bucket. bindDBConn decides; this only
// spares the promotion in the common case.
func (s *UdpServer) checkDBIdAvailable(dbId, pubKey string) error {
	s.dbConnectionMapMutex.Lock()
	defer s.dbConnectionMapMutex.Unlock()

	if s.dbIdHeldLocked(dbId, pubKey) {
		return errDbIdHeld
	}
	return nil
}

// bindDBConn makes conn the route for conn.DBId. It refuses when the dbId
// is held by another nhp-db that is still active. An nhp-db holds one dbId:
// coming online with a new one releases the old one.
func (s *UdpServer) bindDBConn(conn *DBConn) error {
	pubKey := conn.DBPeer.PublicKeyBase64()

	s.dbConnectionMapMutex.Lock()
	defer s.dbConnectionMapMutex.Unlock()

	if s.dbIdHeldLocked(conn.DBId, pubKey) {
		return errDbIdHeld
	}
	for dbId, c := range s.dbConnectionMap {
		if dbId != conn.DBId && c.DBPeer.PublicKeyBase64() == pubKey {
			delete(s.dbConnectionMap, dbId)
		}
	}
	s.dbConnectionMap[conn.DBId] = conn
	return nil
}

// unbindDBConn removes the routes that point at connData, and only those:
// before an old connection times out, the same nhp-db may already be online
// again over a newer one.
func (s *UdpServer) unbindDBConn(connData *core.ConnectionData) {
	s.dbConnectionMapMutex.Lock()
	defer s.dbConnectionMapMutex.Unlock()

	for dbId, c := range s.dbConnectionMap {
		if c.ConnData.Equal(connData) {
			delete(s.dbConnectionMap, dbId)
		}
	}
}

// dbIdHolder returns the public key of the active nhp-db online with dbId,
// or "" if there is none.
func (s *UdpServer) dbIdHolder(dbId string) string {
	s.dbConnectionMapMutex.Lock()
	defer s.dbConnectionMapMutex.Unlock()

	c, found := s.dbConnectionMap[dbId]
	if !found {
		return ""
	}
	pubKey := c.DBPeer.PublicKeyBase64()
	if !s.activeDBPeerLocked(pubKey) {
		return ""
	}
	return pubKey
}

// Why dbConnForObject found no route for an object. Only the server log
// sees these; the agent gets the generic ErrDBOffline either way.
var (
	errDbIdNotOnline      = errors.New("no nhp-db is online with this dbId")
	errDbIdHolderInactive = errors.New("the nhp-db online with this dbId is no longer configured or has expired")
	errDbIdHolderNotOwner = errors.New("the nhp-db online with this dbId is not the object's owner")
)

// dbConnForObject returns the connection to send an object's NHP_DWR on:
// the active nhp-db online with the object's dbId and, when the object has
// a recorded owner, only if that nhp-db is the owner. Otherwise the object
// is treated as offline, so its wrapping request never reaches an nhp-db
// that merely claims its dbId. The error says which case it was;
// errDbIdHolderNotOwner names the holder, because a holder that is not the
// owner is what squatting a dbId looks like.
func (s *UdpServer) dbConnForObject(dbId, owner string) (*DBConn, error) {
	s.dbConnectionMapMutex.Lock()
	defer s.dbConnectionMapMutex.Unlock()

	c, found := s.dbConnectionMap[dbId]
	if !found {
		return nil, errDbIdNotOnline
	}
	pubKey := c.DBPeer.PublicKeyBase64()
	if !s.activeDBPeerLocked(pubKey) {
		return nil, fmt.Errorf("%w (holder %s)", errDbIdHolderInactive, pubKey)
	}
	if owner != "" && pubKey != owner {
		return nil, fmt.Errorf("%w (holder %s, owner %s)", errDbIdHolderNotOwner, pubKey, owner)
	}
	return c, nil
}
