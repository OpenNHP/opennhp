// This file contains the concrete implementation of the policy interface.
// User should focus on implementing the Policy interface methods in this package
// to define their specific policy logic.
//
// After finished, user should import this package in their policy.go file,
// and create the corresponding policy instance to override PolicyInstance
// which is defined in pkg/policy/main/policy.go.
//
// Policy interface is defined in pkg/policy/interface.go

package impl

import (
	"encoding/json"

	"github.com/OpenNHP/opennhp/nhp/core/wasm/policy"
)

type Attestation struct {
	KernelVersion   string `json:"kernelVersion"`
	InKataContainer bool   `json:"inKataContainer"`
	CanDialIn       bool   `json:"canDialIn"`
	CPUModel        string `json:"cpuModel"`
	ProductName 	string `json:"productName"`
	Manufacturer 	string `json:"manufacturer"`
}

type PolicyImpl struct {
}

func NewPolicy() policy.Policy {
	return &PolicyImpl{}
}

func (p *PolicyImpl) OnAttestationCollect() (attestation string) {
	attestationStruct := Attestation{
		KernelVersion:   "4.19.121-linuxkit",
		InKataContainer: true,
		CanDialIn:	   true,
		CPUModel:		"Intel(R) Core(TM) i7-8550U CPU @ 1.80GHz",
		ProductName: 	"Intel(R) Core(TM) i7-8550U CPU @ 1.80GHz",
	}

	attestationJson, _ := json.Marshal(attestationStruct)

	return string(attestationJson)
}

func (p *PolicyImpl) OnAttestationVerify(attestation string) bool {
	policy.Log("OnAttestationVerify: log attestation in wasm vm: " + attestation)

	return true
}

func (p * PolicyImpl) OnDataPreprocess(metadata string, rawData string, filter string) (processedData string) {
	policy.Log("log metadata in wasm vm: " + metadata)
	policy.Log("log rawData in wasm vm: " + rawData)
	policy.Log("log filter in wasm vm: " + filter)

	return "{\"preprocessData\": \"example\"}"
}

func (p * PolicyImpl) OnDataPostprocess(rawOutput string) (processedOutput string) {
	policy.Log("OnDataPostprocess: log rawOutput in wasm vm: " + rawOutput)

	return rawOutput
}
