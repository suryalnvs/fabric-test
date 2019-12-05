module github.com/hyperledger/fabric-test/regression/smoke

go 1.13

replace github.com/hyperledger/fabric-test => ../../../fabric-test

require (
	github.com/hyperledger/fabric-test v1.1.1-0.20191206195025-21b803e98dcd // indirect
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	github.com/pkg/errors v0.8.1 // indirect
)
