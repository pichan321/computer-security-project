# Final Project - Advanced Computer Security (Fall 2024)

## Installations (All required)
1. Go: [Download Go](https://go.dev/doc/install)
  - Please download the right Go according to your operating system
2. Docker: [Download Docker](https://www.docker.com/products/docker-desktop/)
  - Please download the right Docker according to your operating system
3. IPFS: [Download and install IPFS via Docker](https://docs.ipfs.tech/install/run-ipfs-inside-docker/#set-up)
  - Once you have IPFS installed as an image in Docker, please convert it into a running container, then proceed to the next step
  ```
  docker pull ipfs/kubo
  docker run -d --name ipfs -p 4001:4001 -p 127.0.0.1:8080:8080 -p 127.0.0.1:5001:5001 ipfs/kubo
```

## How to run
1. Clone the repository
2. Perform `go get` to install all the required Go modules/packages/dependencies
3. Perform `go test -v ./...` to run the security test suite implemented
  - Supposedly, if everything was installed correctly, all test cases should pass successfully without errors
  - If in any cases, you do encounter a nil error, it might be related to IPFS container issue. Please contact me at anytime if this arises.

