GPtt
==========

Official golang implementation of the Ptt.ai Framework.

The architecture of Ptt.ai can be found in the [link](https://docs.google.com/presentation/d/1q44LYz0i-iMxXMD9zfV9kqwah9UJGFOaQZxs0GvM5E4/edit#slide=id.p) [(中文版)](https://docs.google.com/presentation/d/1X6fGAElPtvsMK8Fys8VwSj9UPfNRkRRHDE0lQcUyK4Y/edit#slide=id.p)

Install
-----

    git clone git@gitlab.corp.ailabs.tw:ptt.ai/go-pttai.git; cd go-pttai;
    make gptt
    build/bin/gptt


Docker Environment
-----

    ./scripts/docker_build.sh
    ./scripts/docker.sh
    ./scripts/docker_stop.sh


Unit-Test
-----

    make test


Testing for specific dir:

    ./scripts/test.sh [dir]
    (ex: ./scripts/test.sh common)


Running Godoc
-----

    ./scripts/doc.sh


Development
-----
The code-structure is based on [go-ethereum](https://github.com/ethereum/go-ethereum). The following is the general guide for the development.

    git clone git@gitlab.corp.ailabs.tw:ptt.ai/go-pttai.git; cd go-pttai; git submodule update --init; ./scripts/init_cookiecutter.sh

    ./scripts/gptt-dev.sh

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
