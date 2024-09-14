#ifdef __cplusplus
extern "C" {
#endif
#ifndef NHPDEVICEDEF_H
#define NHPDEVICEDEF_H

#include <stdlib.h>
#include <string.h>
#include <stdarg.h>

#define NHP_COOKIE_SIZE 32
#define NHP_PUBLIC_KEY_SIZE 64

typedef enum _NhpDeviceType {
    NHP_AGENT = 1,
    NHP_SERVER,
    NHP_AC,
} NhpDeviceType;

typedef enum _NhpCipherScheme {
    NHP_CIPHER_SCHEME_CURVE,
    NHP_CIPHER_SCHEME_GMSM,
} NhpCipherScheme;

typedef enum _NhpMsgType {
    NHP_KPL = 0, // general keepalive packet
	NHP_KNK,        // agent sends knock to server
	NHP_ACK,        // server replies knock status to agent
	NHP_AOP,        // server asks ac for operation
	NHP_ART,        // ac replies server for operation result
	NHP_LST,        // agent requests server for listing services and applications
	NHP_LRT,        // server replies to agent with services and applications result
	NHP_COK,        // server sends cookie to agent
	NHP_RKN,        // agent sends reknock to server
	NHP_RLY,        // relay sends relayed packet to server
	NHP_AOL,        // ac sends online status to server
	NHP_AAK,        // server sends ack to ac after receving ac's online status
	NHP_OTP,        // agent requests server for one-time-password
	NHP_REG,        // agent asks server for registering
	NHP_RAK,        // server sends back ack when agent registers correctly
	NHP_ACC,        // agent sends to ac/resource for actual ip access
	NHP_EXT,        // agent requests immediate disconnection
} NhpMsgType;

typedef struct _NhpResult {
   size_t handle; // The handle of the nhp device returned after initialization
   int errCode;   // Error code: 0 for no error, non-zero values indicate errors
   char *errMsg;  // Error message: NULL for no error, a value indicates an error description
} NhpResult;

typedef struct _NhpEncryptParams {
    // Specifies the encryption scheme used for the message:
    // 0: curve25519/chacha20poly1305/blake2s
    // 1: sm2/sm4/sm3
    unsigned char cipherScheme;
    // true: Use zlib to compress the plaintext message
    unsigned char compress;
    // reserved
    unsigned char reserved0;
    // reserved
    unsigned char reserved1;
    // If responding to a previously received NHP packet, this specifies the last transaction ID of the sender.
    // if set to 0, a new transaction ID is generated locally
    unsigned long long assignTransactionId;
    // The cookie value specified by the remote peer for re-knocking.
    // note: the device has a different cookie for each peer connection
    unsigned char cookie[NHP_COOKIE_SIZE]; 
} NhpEncryptParams;

typedef struct _NhpEncryptResult {
    int errCode;                        // Error code
    int packetLen;                      // Packet length
    unsigned long long transactionId;   // Local transaction ID
    char *errMsg;                       // Error message
    unsigned char *packet;              // Packet data
} NhpEncryptResult;

typedef struct _NhpPubicKey {
    unsigned char data[NHP_PUBLIC_KEY_SIZE];
} NhpPubicKey;

typedef struct _NhpCookieStore {
    unsigned char currCookie[NHP_COOKIE_SIZE];
    unsigned char prevCookie[NHP_COOKIE_SIZE];
    long long lastCookieTime;
} NhpCookieStore;

typedef struct _NhpConnContext {
    NhpPubicKey *peerPbk;
    NhpCookieStore *cookieStore;
    long long *lastPeerSendTime;
} NhpConnContext;

typedef struct _NhpDecryptResult {
    int errCode;                            // Error code
    int msgType;                            // Message type
    int msgIdLen;                           // Sender identification length
    int dataLen;                            // Payload length
    unsigned long long msgTransactionId;    // Message sender transaction ID
    char *errMsg;                           // Error description
    unsigned char *msgId;                   // Message sender identification
    unsigned char *data;                    // Payload data
} NhpDecryptResult;

typedef enum _NhpError {
    // general
    ERR_NHP_SUCCESS = 0,
    ERR_NHP_DEVICE_NOT_INITIALIZED = 30000,
    ERR_NHP_DEVICE_ALREADY_CREATED,
    ERR_NHP_CIPHER_NOT_SUPPORTED,
    ERR_NHP_OPERATION_NOT_APPLICABLE,
    ERR_NHP_CREATE_DEVICE_FAILED,
    ERR_NHP_CLOSE_DEVICE_FAILED,
    ERR_NHP_SDK_RUNTIME_PANIC,

    // encryption
    ERR_NHP_EMPTY_PEER_PUBLIC_KEY = 31001,
    ERR_NHP_EPHERMAL_ECDH_PEER_FAILED,
    ERR_NHP_DEVICE_ECDH_PEER_FAILED,
    ERR_NHP_IDENTITY_TOO_LONG,
    ERR_NHP_DATA_COMPRESSION_FAILED,
    ERR_NHP_PACKET_SIZE_EXCEEDS_BUFFER,

    // decryption
    ERR_NHP_CLOSE_CONNECTION = 32001,
    ERR_NHP_INCORRECT_PACKET_SIZE,
    ERR_NHP_MESSAGE_TYPE_NOT_MATCH_DEVICE,
    ERR_NHP_SERVER_OVERLOAD,
    ERR_NHP_HMAC_CHECK_FAILED,
    ERR_NHP_SERVER_HMAC_CHECK_FAILED,
    ERR_NHP_DEVICE_ECDH_EPHERMAL_FAILED,
    ERR_NHP_PEER_IDENTITY_VERIFICATION_FAILED,
    ERR_NHP_AEAD_DECRYPTION_FAILED,
    ERR_NHP_DATA_DECOMPRESSION_FAILED,
    ERR_NHP_DEVICE_ECDH_OBTAINED_PEER_FAILED,
    ERR_NHP_SERVER_REJECT_WITH_COOKIE,
    ERR_NHP_REPLAY_PACKET_RECEIVED,
    ERR_NHP_FLOOD_PACKET_RECEIVED,
    ERR_NHP_STALE_PACKET_RECEIVED,

} NhpError;


#endif // NHPDEVICEDEF_H
#ifdef __cplusplus
}
#endif
