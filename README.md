GPtt
==========

Official golang implementation of the Ptt.ai Framework.

[![API Reference](https://godoc.org/github.com/ailabstw/go-pttai?status.png)](https://godoc.org/github.com/ailabstw/go-pttai)
[![Travis](https://travis-ci.org/ailabstw/go-pttai.svg?branch=master)](https://travis-ci.org/ailabstw/go-pttai)

The architecture of Ptt.ai can be found in the [link](https://docs.google.com/presentation/d/1q44LYz0i-iMxXMD9zfV9kqwah9UJGFOaQZxs0GvM5E4/edit#slide=id.p) [(中文版)](https://docs.google.com/presentation/d/1X6fGAElPtvsMK8Fys8VwSj9UPfNRkRRHDE0lQcUyK4Y/edit#slide=id.p)

More documents can be found in [PIPs](https://github.com/ailabstw/PIPs)

Install
-----

    go get github.com/ailabstw/go-pttai
    go install github.com/ailabstw/go-pttai/cmd/gptt
    ${GOPATH}/bin/gptt


Docker Environment
-----

    ./scripts/docker_build.sh
    ./scripts/docker.sh
    ./scripts/docker_stop.sh


Docker with customized storage / ports
-----

    ./sciprts/docker_with_storage.sh -p [http-port in local-machine] -a [api-port in local-machine] -q [p2p-port in local-machine] -e [optional external http/https addr] -b [optional external api addr] -s [dir]


Unit-Test
-----

    make test


Testing for specific dir:

    ./scripts/test.sh [dir]
    (ex: ./scripts/test.sh common)

E2E-Test
-----

Testing for specific e2e-test

    ./scripts/e2e.sh [any test in the e2e dir]

ex: ./scripts/e2e.sh friend_basic runs TestFriendBasic


Running Godoc
-----

    ./scripts/doc.sh


Development
-----
The code-structure is based on [go-ethereum](https://github.com/ethereum/go-ethereum). The following is the general guide for the development.

    go get github.com/ailabstw/go-pttai; cd ${GOPATH}/src/github.com/ailabstw/go-pttai; ./scripts/init_cookiecutter.sh; source __/bin/activate

    ./scripts/gptt-testnet.sh

* follow gofmt / goimports
* follow [gotests](https://github.com/cweill/gotests)
* coding style (in-general):
    1. Each struct file (generated from ./scripts/dev_struct.sh) represents 1 struct
    2. Each module file (generated from ./scripts/dev_module.sh) represents 1 public function
    3. Struct is always CapitalizedCamelCase
    4. Public constants / variables / functions / member variables / member functions are CapitalizedCamelCase
    5. local constants / variables / functions / member variables / member functions are lowerCamelCase
    6. Global variables are in globals.go
    7. Global test variables are in globals_test.go
    8. Errors are in errors.go
* Naming:
    * Full name: Pttai, pttai
    * cmd: gptt, Gptt (go-ptt)
* Default ports:
    * http-connection: 9774
    * api-connection: 14779
    * p2p-connection: 9487
