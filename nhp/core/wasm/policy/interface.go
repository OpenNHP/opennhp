package policy

// Policy interface implemented in Go, designed for interaction between host and a WebAssembly (WASM) virtual machine.
// That means that caller (here called engine) implements this interface to interact with WebAssembly (WASM) virtual machine.
// Policy provider implments this interface for concrete policy.
// It is intended to be language-agnostic: equivalent interfaces should be implemented in other
// languages (e.g., Rust, C, AssemblyScript) to support the same behavior in cross-platform WASM environments.

// Notes: the format of all string parameters is JSON.

type Policy interface {
	// Collect attestation, policy provider needs to implement this method to collect any required attestation
	OnAttestationCollect() (attestation string)

	// Verify attestation, policy provider needs to implement this method to verify attestation which is collected by OnAttestationCollect
	OnAttestationVerify(attestation string) bool

	// Preprocess data. The policy provider must implement this method to transform or sanitize the data, ensuring it is safe and compliant for use by the data consumer.
	OnDataPreprocess(metadata string, rawData string, filter string) (processedData string)

	// Postprocess data. The policy provider must implement this method to validate data returned by the data consumer, ensuring it remains safe and policy-compliant.
	OnDataPostprocess(rawOutput string) (processedOutput string)
}
