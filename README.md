# Final Project - Advanced Computer Security (Fall 2024)

## Installations (All required)
1. Go: [Download Go](https://go.dev/doc/install)
  - Please download the right Go according to your operating system
2. Docker: [Download Docker](https://www.docker.com/products/docker-desktop/)
  - Please download the right Docker according to your operating system
3. IPFS:
  - Once you have Docker installed, please have it running as a daemon
  - Install IPFS via Docker: [](https://docs.ipfs.tech/install/run-ipfs-inside-docker/#set-up)

## How to run
1. Clone the repository
2. Perform `go get` to install all the required Go modules/packages/dependencies
3. Perform `go test -v ./...` to run the security test suite implemented
  - Supposedly, if everything was installed correctly, all test cases should pass successfully without errors
  - If in any cases, you do encounter a nil error, it might be related to IPFS container issue. Please contact me at anytime if this arises.

