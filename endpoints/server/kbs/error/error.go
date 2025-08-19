/*
Package error defines error types and utility functions for the KBS (Key Broker Service)
component. It provides standardized error responses with specific error types and
detail messages for common failure scenarios encountered in the KBS operations,
including token validation, resource access, encryption processes, and session management.
*/

package error


type ErrorType string

const (
	tokenNotFound             ErrorType = "https://github.com/confidential-containers/kbs/errors/TokenNotFound"
	tokenInvalid              ErrorType = "https://github.com/confidential-containers/kbs/errors/TokenInvalid"
	policyDeny                ErrorType = "https://github.com/confidential-containers/kbs/errors/PolicyDeny"
	teePubKeyNotFound         ErrorType = "https://github.com/confidential-containers/kbs/errors/TeePubKeyNotFound"
	resourceNotFound          ErrorType = "https://github.com/confidential-containers/kbs/errors/ResourceNotFound"
	keyGenerationFailed       ErrorType = "https://github.com/confidential-containers/kbs/errors/KeyGenerationFailed"
	keyEncryptionFailed       ErrorType = "https://github.com/confidential-containers/kbs/errors/KeyEncryptionFailed"
	contentEncryptionFailed   ErrorType = "https://github.com/confidential-containers/kbs/errors/ContentEncryptionFailed"
	invalidRequest            ErrorType = "https://github.com/confidential-containers/kbs/errors/InvalidRequest"
	nonceGenerationFailed     ErrorType = "https://github.com/confidential-containers/kbs/errors/NonceGenerationFailed"
	sessionIDGenerationFailed ErrorType = "https://github.com/confidential-containers/kbs/errors/SessionIDGenerationFailed"
	missingOrInvalidSessionID ErrorType = "https://github.com/confidential-containers/kbs/errors/MissingOrInvalidSessionID"
	tokenGenerationFailed     ErrorType = "https://github.com/confidential-containers/kbs/errors/TokenGenerationFailed"
)

func newError(errType ErrorType, detail string) map[string]any {
	return map[string]any{
		"type":   errType,
		"detail": detail,
	}
}

func TokenNotFound() map[string]any {
	return newError(tokenNotFound, "Attestation Token is not found")
}

func TokenInvalid() map[string]any {
	return newError(tokenInvalid, "Attestation Token is invalid")
}

func PolicyDeny() map[string]any {
	return newError(policyDeny, "Access is denied by policy")
}

func TeePubKeyNotFound(err error) map[string]any {
	return newError(teePubKeyNotFound, "Tee public key is not found"+err.Error())
}

func ResourceNotFound(err error) map[string]any {
	return newError(resourceNotFound, "Resource is not found: "+err.Error())
}

func KeyGenerationFailed(err error) map[string]any {
	return newError(keyGenerationFailed, "Failed to generate content key: "+err.Error())
}

func KeyEncryptionFailed(err error) map[string]any {
	return newError(keyEncryptionFailed, "Failed to encrypt content key: "+err.Error())
}

func ContentEncryptionFailed(err error) map[string]any {
	return newError(contentEncryptionFailed, "Failed to encrypt resource content: "+err.Error())
}

func InvalidRequest(err error) map[string]any {
	return newError(invalidRequest, "Invalid request: "+err.Error())
}

func NonceGenerationFailed(err error) map[string]any {
	return newError(nonceGenerationFailed, "Failed to generate nonce: "+err.Error())
}

func SessionIDGenerationFailed(err error) map[string]any {
	return newError(sessionIDGenerationFailed, "Failed to generate session ID: "+err.Error())
}

func MissingOrInvalidSessionID() map[string]any {
	return newError(missingOrInvalidSessionID, "Missing or invalid session ID")
}

func TokenGenerationFailed(err error) map[string]any {
	return newError(tokenGenerationFailed, "Failed to generate token: "+err.Error())
}
