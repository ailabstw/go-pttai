gptt
==========

Official golang implementation of the PTT.ai Framework.

[![API Reference](https://godoc.org/github.com/ailabstw/go-pttai?status.png)](https://godoc.org/github.com/ailabstw/go-pttai)
[![Travis](https://travis-ci.org/ailabstw/go-pttai.svg?branch=master)](https://travis-ci.org/ailabstw/go-pttai)

The architecture of PTT.ai can be found in the [link](https://docs.google.com/presentation/d/1q44LYz0i-iMxXMD9zfV9kqwah9UJGFOaQZxs0GvM5E4/edit#slide=id.p) [(中文版)](https://docs.google.com/presentation/d/1X6fGAElPtvsMK8Fys8VwSj9UPfNRkRRHDE0lQcUyK4Y/edit#slide=id.p)

More documents can be found in [PIPs](https://github.com/ailabstw/PIPs)

Install
-----

    go get github.com/ailabstw/go-pttai
    go install github.com/ailabstw/go-pttai/cmd/gptt
    ${GOPATH}/bin/gptt


License
-----

All License can be found in docs/


Docker Environment
-----

    ./scripts/docker_build.sh
    ./scripts/docker.sh
    ./scripts/docker_stop.sh


Docker with customized storage / ports
-----

    ./scripts/docker_with_storage.sh -p [http-port in local-machine] -a [api-port in local-machine] -q [p2p-port in local-machine] -e [optional external http/https addr] -b [optional external api addr] -s [dir]


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
    * Full name: PTTai, pttai, PTT.ai
    * cmd: gptt, Gptt (go-ptt)
* Default ports:
    * http-connection: 9774
    * api-connection: 14779
    * p2p-connection: 9487


SSH Tunnel
-----

1. ssh -L [rpcport-for-localhost]:localhost:[rpcport-for-remote-host] [username]@[remote-host]
2. ssh -L [httpport-for-localhost]:localhost:[httpport-for-remote-host] [username]@[remote-host]
3. ./build/bin/gptt --testp2p --rpcport [port-for-remote-host] --httpaddr 127.0.0.1:[http-port] --exthttpaddr http://localhost:[httpport-for-localhost] --extrpcaddr http://localhost:[rpcport-for-localhost]
4. (in browser) http://localhost:[httpport-for-localhost]


Linode
-----

1. Create Linode with Stackscript ID: 399252 (hsiaochuanheng / p.me-docker (go-pttai))
2. Launch the console and ensure that the new linode is created successfully.
3. ssh -L 14779:localhost:14779 admin@[IP]
4. ssh -L 9774:localhost:9774 admin@[IP]
5. open browser and connect to http://localhost:9774
6. Update (in admin@[IP]):

    ```
    sudo -s
    docker pull ailabstw/go-pttai:latest
    HTTPPORT=9774
    APIPORT=14779
    docker run -e HTTPPORT=${HTTPPORT} -e APIPORT=${APIPORT} -itd --restart=always -p 9487:9487 -p 127.0.0.1:9774:9774 -p 127.0.0.1:14779:14779 -v /home/admin/pttai.docker:/root/.pttai --name go-pttai ailabstw/go-pttai:latest gptt "--testp2p" "--httpdir" "/static" "--httpaddr" "0.0.0.0:9774" "--rpcaddr" "0.0.0.0" "--exthttpaddr" "http://localhost:${HTTPPORT}" "--extrpcaddr" "http://localhost:${APIPORT}"
    ```

Digital Ocean
-----

1. Create docker-with-ubuntu in One-click Applications.
2. Install / Update (in root@[IP]):

    ```
    sudo -s
    docker pull ailabstw/go-pttai:latest
    HTTPPORT=9774
    APIPORT=14779
    docker run -e HTTPPORT=${HTTPPORT} -e APIPORT=${APIPORT} -itd --restart=always -p 9487:9487 -p 127.0.0.1:9774:9774 -p 127.0.0.1:14779:14779 -v /home/admin/pttai.docker:/root/.pttai --name go-pttai ailabstw/go-pttai:latest gptt "--testp2p" "--httpdir" "/static" "--httpaddr" "0.0.0.0:9774" "--rpcaddr" "0.0.0.0" "--exthttpaddr" "http://localhost:${HTTPPORT}" "--extrpcaddr" "http://localhost:${APIPORT}"
    ```

3. ssh -L 14779:localhost:14779 root@[IP]
4. ssh -L 9774:localhost:9774 root@[IP]
5. open browser and connect to http://localhost:9774


AWS
-----


1. Setup **Key Pair** in [AWS ec2 console](https://console.aws.amazon.com/ec2/home#KeyPairs). (for SSH use)
2. Setup **Security Group** in [AWS ec2 console](https://console.aws.amazon.com/ec2/home#SecurityGroups)
    * Be sure to have port 22, 9487, 9774, 14779 open to 0.0.0.0/0
3. Create an [ECS cluster](https://console.aws.amazon.com/ecs) with the correpsonding Key Pair and Security Group.
4. Install / Update (in ec2-user@[IP]):

    ```
    sudo -s
    docker pull ailabstw/go-pttai:latest
    HTTPPORT=9774
    APIPORT=14779
    docker run -e HTTPPORT=${HTTPPORT} -e APIPORT=${APIPORT} -itd --restart=always -p 9487:9487 -p 127.0.0.1:9774:9774 -p 127.0.0.1:14779:14779 -v /home/admin/pttai.docker:/root/.pttai --name go-pttai ailabstw/go-pttai:latest gptt "--testp2p" "--httpdir" "/static" "--httpaddr" "0.0.0.0:9774" "--rpcaddr" "0.0.0.0" "--exthttpaddr" "http://localhost:${HTTPPORT}" "--extrpcaddr" "http://localhost:${APIPORT}"
    ```

5. ssh -L 14779:localhost:14779 ec2-user@[IP]
6. ssh -L 9774:localhost:9774 ec2-user@[IP]
7. open browser and connect to http://localhost:9774


AWS Elastic Beanstalk (On-click deploy)
-----

1. Setup **Key Pair** in [AWS ec2 console](https://console.aws.amazon.com/ec2/home#KeyPairs). (for SSH use)

    Please remeber where you save the private key in your computer, you will need it at the last step.

2. [Click this link](https://console.aws.amazon.com/elasticbeanstalk/?region=ap-northeast-1#/newApplication?applicationName=My.P.Me&platform=Multi-container%20Docker&sourceBundleUrl=https%3A%2F%2Fs3-ap-northeast-1.amazonaws.com%2Fptt.ai%2Fpme.zip&environmentType=SingleInstance&tierName=WebServer&instanceType=m1.small)

    *This will deploy the pre-build [p.me docker image](https://hub.docker.com/r/secret104278/pme) for you*

3. Click choice option `Upload your code`

    (Note: the code is already uploaded for you, you only have to click this option. This seems to be a bug of AWS.)

![](https://i.imgur.com/UDcDEIJ.png)

4. Click button `Review and Launch`

5. Click `Modify` in the `Security` section
![](https://i.imgur.com/yqiKjWw.png)

6. Chose the key pair you just setup in `EC2 key pair`，then `Save`

7. Click button `Create app`

8. ssh -L 14779:localhost:14779 ec2-user@[IP]
9. ssh -L 9774:localhost:9774 ec2-user@[IP]
10. open browser and connect to http://localhost:9774