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
    size_t handle; // 初始化后返回的nhp device句柄
    int errCode; // 错误代码：无错误则为0，其它非0值表示错误
    char *errMsg; // 错误描述：为空则无错误，有值表示错误描述语句
} NhpResult;

typedef struct _NhpEncryptParams {
    unsigned char cipherScheme; // 指定消息采用的加密方案 0: curve25519/chacha20poly1305/blake2s, 1: sm2/sm4/sm3
    unsigned char compress; // true: 使用zlib压缩明文消息
    unsigned char reserved0;
    unsigned char reserved1;
    unsigned long long assignTransactionId; // 如果要响应之前接收的NHP包，则需要指定使用发送方上次的流水号， 值为0则不指定，由本地生成新的流水号
    unsigned char cookie[NHP_COOKIE_SIZE]; // 对端peer要求重新敲门所指定的cookie值，注：device在每个peer连接下具有不同的cookie
} NhpEncryptParams;

typedef struct _NhpEncryptResult {
    int errCode; // 错误代码
    int packetLen; // 报文长度
    unsigned long long transactionId; // 本地交易流水号
    char *errMsg; // 错误描述
    unsigned char *packet; // 报文数据
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
    int errCode; // 错误代码
    int msgType; // 消息类型
    int msgIdLen; // 发送方标识长度
    int dataLen; // 载荷长度
    unsigned long long msgTransactionId; // 消息发送方交易流水号
    char *errMsg; // 错误描述
    unsigned char *msgId; // 消息发送方标识
    unsigned char *data; // 载荷数据
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
