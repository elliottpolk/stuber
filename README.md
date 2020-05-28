# Stuber

## Why

Stuber /st(y)o͞o·ber/ is a configurable stubbing tool meant to provide
mocked stub services, typically for testing.

For configuring, have a look at the [sample.json](/test_data/sample.json) file for the layout and options available. **Note** that `stuber` defaults to the `./data` directory for loading mocks.

---

## Builds
The service currently uses the **(optional)** tool ```make``` for builds. The _golang Docker image_ has this pre-installed. To build:

```bash
# ensure the below commands will be run from the project dir
$ cd $GOPATH/src/git.platform.manulife.io/gwam/stuber

# the default targets are 'clean build'
$ make

# to build and package for deployment to PCF or Artifactory
$ make clean build package

# using the Go Docker image
$ docker run --rm -it -v "${GOPATH}":/go -w /go/src/git.platform.manulife.io/gwam/${PWD##*/} golang:latest /bin/sh -c 'make'

# to build and package for deployment to PCF or Artifactory
$ docker run --rm -it -v "${GOPATH}":/go -w /go/src/git.platform.manulife.io/gwam/${PWD##*/} golang:latest /bin/sh -c 'make clean build package'

```

**NOTE:** the default build is for **_Linux_** environments. The ```GOOS=<environment>``` must be specified for OS-specific builds.

```bash

# macOS
$ GOOS=darwin make
# OR
$ docker run --rm -it -v "${GOPATH}":/go -w /go/src/git.platform.manulife.io/gwam/${PWD##*/} golang:latest /bin/sh -c 'GOOS=darwin make'

# Windows
$ GOOS=windows make
# OR
$ docker run --rm -it -v "${GOPATH}":/go -w /go/src/git.platform.manulife.io/gwam/${PWD##*/} golang:latest /bin/sh -c 'GOOS=windows make'
```

---

### Testing

Testig is straight forward. It can be run with the following command:

```bash

# unit tests
$ make clean test

```

---

## Running

Running `stuber` is also pretty straight forward, especially since it's just a simple RESTful server.

```bash
$ stuber -h

NAME:
   stuber - A new cli application

USAGE:
   stuber [global options] command [command options] [arguments...]

VERSION:
   0.1.0

DESCRIPTION:
   Stuber /st(y)o͞o·ber/ is a configurable stubbing tool meant to provide mocked stub services, typically for testing 🧪

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config-file value, -c value, --cfg value, --confg value, --config value  optional path to config file
   --http-port value                                                          HTTP port to listen on (default: "8080") [$STUBER_HTTP_PORT]
   --tls-port value                                                           HTTPS port to listen on (default: "8443") [$STUBER_HTTPS_PORT]
   --tls-cert value                                                           TLS certificate file for HTTPS [$STUBER_TLS_CERT]
   --tls-key value                                                            TLS key file for HTTPS [$STUBER_TLS_KEY]
   --data-dir value, -d value, --dir value, --data value                      data directory for stub JSON files (default: "./data")
   --help, -h                                                                 show help (default: false)
   --version, -v                                                              print the version (default: false)

COPYRIGHT:
   Copyright © 2018-2020 Elliott Polk

$ stuber -d './some-mock-data-dir'
```

## Example

```bash

# start the server using the sample data
$ stuber -d ./test_data
DEBU[0000] processing files for dir ./test_data
DEBU[0000] processing file test_data/sample.json
DEBU[0000] adding handlers for route /api/v1/tests/dummy
INFO[0000] HTTP listening on port 8080

# ---
# in another terminal, it can be tested by using curl
$ curl -s -S --fail 'localhost:8080/api/v1/tests/dummy' | jq
[
  {
    "email": "acaven0@tinypic.com",
    "first_name": "Alleyn",
    "gender": "Male",
    "id": 1,
    "last_name": "Caven",
    "ssn": "280-74-4151"
  },
  {
    "email": "cpieche1@cmu.edu",
    "first_name": "Christal",
    "gender": "Female",
    "id": 2,
    "last_name": "Pieche",
    "ssn": "239-77-2514"
  },
  {
    "email": "tduckels2@miibeian.gov.cn",
    "first_name": "Taber",
    "gender": "Male",
    "id": 3,
    "last_name": "Duckels",
    "ssn": "872-94-0816"
  },
  {
    "email": "ssnedker3@deviantart.com",
    "first_name": "Sigfried",
    "gender": "Male",
    "id": 4,
    "last_name": "Snedker",
    "ssn": "816-39-2640"
  },
  {
    "email": "fhandforth4@ehow.com",
    "first_name": "Freddie",
    "gender": "Male",
    "id": 5,
    "last_name": "Handforth",
    "ssn": "411-53-8632"
  }
]

```
