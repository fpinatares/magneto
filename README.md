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

## Lambda configuration ##
For the lambda with the function to detect mutans, we have to set 2 environment variables:
* NECESSARY_SEQUENCE (Which for what the requirements says it is 4 by now)
* NECESSARY_SEQUENCES (Which for what the requirements says it is 2 by now)

## Test ##

To run tests we should execute the following command within each of the packages:
```bash
go test -cover
```
__NOTE:__ The cover flag allows to see the code coverage within the package

## How to use ##

### API ###
#### Analysis ####
To Analize a certain DNA, a POST request should be made to the following endpoint 
If the DNA provided is from a Mutant, the endpoint will return 200 - OK
If the DNA provided is from a Human, the endpoint will return 403 - Forbidden
If the DNA provided is malformed, the endpoint will return a 400 - Bad Request

Header       | Value
------------ | -------------
Content-Type | application/json

```
POST https://rhpbk7pt2m.execute-api.us-east-1.amazonaws.com/v1/mutant
```
Body example with a Mutant DNA:
```json
{
    "dna":["ATGCGA","CAGTGC","TTATGT","AGAAGG","CCCCTA","TCACTG"]
}
```
Body example with a Human DNA:
```json
{
    "dna":["CCCACC", "CAGTGC", "TTATTT", "AGACGG", "GCGTCA", "TCACTG"]
}
```
__NOTE:__ Each string should only be a combination of the followings 4 letters, otherwise it will be considered malformed: A (Adenanina), C (Citosina), G (Guanina), T (Timina)

#### Statistics ####
To get the statistics, a GET request should be made to the following endpoint
```
GET https://rhpbk7pt2m.execute-api.us-east-1.amazonaws.com/v1/stats
```
__NOTE:__ The ratio is rounded to 2 decimal points. If the count of Humans is 0, the ratio will show the count of Mutants.


