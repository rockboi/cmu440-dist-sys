p1
==

This repository contains the starter code for project 1 (15-440, Fall 2017). It also contains
the tests that we will use to grade your implementation, and two simple echo server/client
(`srunner` and `crunner`, respectively) programs that you might find useful for your own testing
purposes. These instructions assume you have set your `GOPATH` to point to the repository's
root `p1/` directory.

If at any point you have any trouble with building, installing, or testing your code, the article
titled [How to Write Go Code](http://golang.org/doc/code.html) is a great resource for understanding
how Go workspaces are built and organized. You might also find the documentation for the
[`go` command](http://golang.org/cmd/go/) to be helpful. As always, feel free to post your questions
on Piazza.

This project was designed for, and tested on AFS cluster machines, though you may choose to
write and build your code locally as well.

## Part A

### Testing your code using `srunner` & `crunner`

To make testing your server a bit easier we have provided two simple echo server/client
programs called `srunner` and `crunner`. If you look at the source code for the two programs,
you’ll notice that they import the `github.com/cmu440/lsp` package (in other words, they compile
against the current state of your LSP implementation). We believe you will find these programs
useful in the early stages of development when your client and server implementations are
largely incomplete.

To compile, build, and run these programs, use the `go run` command from inside the directory
storing the file (these instructions assume your `GOPATH` is pointing to the project’s root
`p1/` directory):

```bash
go run srunner.go
```

The `srunner` and `crunner` programs may be customized using command line flags. For more
information, specify the `-h` flag at the command line. For example,

```bash
$ go run srunner.go -h
Usage of bin/srunner:
  -elim=5: epoch limit
  -ems=2000: epoch duration (ms)
  -port=9999: port number
  -rdrop=0: network read drop percent
  -v=false: show srunner logs
  -wdrop=0: network write drop percent
  -wsize=1: window size
  -maxBackoff: maximum interval epoch
```

### Running the tests

To test your submission, we will execute the following command from inside the
`p1/src/github.com/cmu440/lsp` directory for each of the tests (where `TestName` is the
name of one of the 44 test cases, such as `TestBasic6` or `TestWindow1`):

```sh
go test -run=TestName
```

Note that we will execute each test _individually_ using the `-run` flag and by specify a regular expression
identifying the name of the test to run. To ensure that previous tests don’t affect the outcome of later tests,
we recommend executing the tests individually (or in small batches, such as `go test -run=TestBasic` which will
execute all tests beginning with `TestBasic`) as opposed to all together using `go test`.

On some tests, we will also check your code for race conditions using Go’s race detector:

```sh
go test -race -run=TestName
```

We have also provided Autolab test scripts mocks in `sh/`. When you are inside the
`p1/src/github.com/cmu440/lsp` directory and execute corresponding script, you can have a rough sense of what your
score should be like on Autolab.

### Submitting to Autolab

As with project 0, we will be using Autolab to grade your submissions for this project.
We will run some&mdash;but not all&mdash;of the tests with the race detector enabled.

To submit your code to Autolab, create a `lsp.tar` file containing your LSP implementation as follows:

```sh
cd p1/src/github.com/cmu440/
tar -cvf lsp.tar lsp/
```

## Part B

### Importing the `bitcoin` package

In order to use the starter code we provide in the `hash.go` and `message.go` files, use the
following `import` statement:

```go
import "github.com/cmu440/bitcoin"
```

Once you do this, you should be able to make use of the `bitcoin` package as follows:

```go
hash := bitcoin.Hash("thom yorke", 19970521)

msg := bitcoin.NewRequest("jonny greenwood", 200, 71010)
```

### Compiling the `client`, `miner` & `server` programs

To compile the `client`, `miner`, and `server` programs, use the `go install` command
as follows (these instructions assume your
`GOPATH` is pointing to the project's root `p1/` directory):

```bash
# Compile the client, miner, and server programs. The resulting binaries
# will be located in the $GOPATH/bin directory.
go install github.com/cmu440/bitcoin/client
go install github.com/cmu440/bitcoin/miner
go install github.com/cmu440/bitcoin/server

# Start the server, specifying the port to listen on.
$GOPATH/bin/server 6060

# Start a miner, specifying the server's host:port.
$GOPATH/bin/miner localhost:6060

# Start the client, specifying the server's host:port, the message
# "bradfitz", and max nonce 9999.
$GOPATH/bin/client localhost:6060 bradfitz 9999
```

Note that you will need to use the `os.Args` variable in your code to access the user-specified
command line arguments.

### Run Sanity Tests

We have provided *basic* tests for your miner and client implementations. Note that passing them does not indicate that your implementation is correct, nor does it mean your code will earn full scores on Autolab. Extra tests are encouraged before you submit your code.

To sanity tests, you need to ensure you have compiled version of `client`, `miner` and `server` in `$GOPATH/bin`. Then you can run `ctest` and `mtest` (without any parameter) in `$GOPATH/bin/{YOUR-OS}/`.

### Submitting to Autolab

To submit your code to Autolab, create a `cmu440.tar` file containing your part A and part B implementation
as follows:

```sh
cd p1/src/github.com/
tar -cvf cmu440.tar cmu440/
```

## Miscellaneous

### Reading the API Documentation

Before you begin the project, you should read and understand all of the starter code we provide.
To make this experience a little less traumatic (we know, it's a lot :P),
fire up a web server and read the documentation in a browser by executing the following command:

```sh
godoc -http=:6060 &
```

Then, navigate to [localhost:6060/pkg/github.com/cmu440](http://localhost:6060/pkg/github.com/cmu440) in a browser.
Note that you can execute this command from anywhere in your system (assuming your `GOPATH`
is pointing to the project's root `p1/` directory).
