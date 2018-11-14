#!/bin/bash

go test -race -run TestBasic1
go test -race -run TestBasic2
go test -race -run TestBasic3
go test -run TestBasic4
go test -run TestBasic5
go test -run TestBasic6
go test -run TestBasic7
go test -run TestBasic8
go test -run TestBasic9
go test -run TestOutOfOrderMsg1
go test -run TestOutOfOrderMsg2
go test -run TestOutOfOrderMsg3
