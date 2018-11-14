# Project 2: The Raft Consensus Algorithm

This repository contains the starter code for project 2 (15-440/15-640, Fall 2018).
These instructions assume you have set your `GOPATH` to point to the repository's
root `p2/` directory.

## Starter Code

The starter code for this project is organized roughly as follows:

```

src/github.com/cmu440-F18/        
  raft/                            Raft implementation, tests and test helpers

  rpc/                             RPC library that must be used for implementing Raft
    
```

## Instructions

##### How to Write Go Code

If at any point you have any trouble with building, installing, or testing your code, the article
titled [How to Write Go Code](http://golang.org/doc/code.html) is a great resource for understanding
how Go workspaces are built and organized. You might also find the documentation for the
[`go` command](http://golang.org/cmd/go/) to be helpful. As always, feel free to post your questions
on Piazza.

### Executing the official tests

#### 1. Checkpoint

To run the checkpoint tests, run the following from the src/github.com/cmu440-F18/raft/ folder

```bash
go test -run 2A
```

#### 2. Full test

To execute all the tests, run the following from the src/github.com/cmu440-F18/raft/ folder

```bash
go test
```

## Miscellaneous

### Using Go on AFS

For those students who wish to write their Go code on AFS (either in a cluster or remotely), you will
need to set the `GOROOT` environment variable as follows (this is required because Go is installed
in a custom location on AFS machines):

```bash
export GOROOT=/usr/local/depot/go
```
