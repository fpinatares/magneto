# Magneto
This repository contains the code and versions for the magneto API with its different components
It contains the code for 3 different AWS Lambda functions 

## Requirements ##
Go 1.15 or higher

## Build ##
Run the followings commands within the root of the project to set the GOARCH and GOOS environment variables and building the packages

### For MacOSx ###
```bash
GOARCH=amd64 GOOS=linux go build mutant.go
```
```bash
GOARCH=amd64 GOOS=linux go build storage/save.go
```
```bash
GOARCH=amd64 GOOS=linux go build stats/stat.go
```

In order to upload them to the lambda functions we should zip them
```bash
zip magneto-mutant.zip mutant
```
```bash
zip magneto-save.zip save
```
```bash
zip magneto-stats.zip stat
```
## Test ##

To run tests we should execute the following command within each of the packages:
```bash
go test -cover
```
__NOTE:__ The cover flag allows to see the code coverage within the package

## How to use ##



