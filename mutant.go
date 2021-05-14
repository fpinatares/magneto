package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/google/uuid"
)

const NECESSARY_SECUENCE = 4
const NECESSARY_SECUENCES = 2
const STATS_TABLE = "stats"
const DNAS_TABLE = "dnas"

//const MUTANT = "Mutant"
//const HUMAN = "Human"

var EnumDnaType = DnaTypes()

func DnaTypes() *DnaType {
	return &DnaType{
		Human:  "Human",
		Mutant: "Mutant",
	}
}

type DnaType struct {
	Human  string
	Mutant string
}

/*
CHEQUEAR QUE LOS TIPOS QUE VENGAN EN EL ADN SEAN DE ESAS 4 LETRAS
HACER TESTS
CONTROL DE EXCEPCIONES
LAMBDA
API GATEWAY
TENGO EL NECESSARY_SECUENCE COMO CONST
documentar lo que devolvemos por default si no llega un post o si la request esta mal
Podria poner el read, parse y validate todo junto en una sola funcion
adenina (A), citosina (C), guanina (G) y timina (T)
asumo que el ratio se calcula mutantes dividido humanos
asumo que para decir si es una secuencia diagonal, cuento a
partir desde donde estoy parado sin mirar atras
validar tamaño de array que todos los strings sean del mismo largo y length del array?
*/

type DnaData struct {
	Uuid string   `json:"uuid"`
	Dna  []string `json:"dna"`
	Type string   `json:"type"`
}

func main() {
	lambda.Start(DetectMutant)
}

func DetectMutant(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.Headers["content-type"] != "application/json" && req.Headers["Content-Type"] != "application/json" {
		return Respond(http.StatusNotAcceptable)
	}
	dnaData, err := ParseRequest(req.Body)
	if err != nil {
		return Respond(http.StatusBadRequest)
	}
	dnaData.Uuid = uuid.New().String()
	dnaData.Type = GetDnaType(dnaData.Dna)
	svc := GetDynamoDBClient()
	err = UpdateData(dnaData, svc)
	if err != nil {
		return Respond(http.StatusInternalServerError)
	}
	if dnaData.Type == EnumDnaType.Mutant {
		return Respond(http.StatusOK)
	}
	return Respond(http.StatusForbidden)
}

func GetDnaType(dna []string) string {
	if IsMutant(dna) {
		return EnumDnaType.Mutant
	} else {
		return EnumDnaType.Human
	}
}

func ParseRequest(body string) (DnaData, error) {
	dnaData := new(DnaData)
	err := json.Unmarshal([]byte(body), &dnaData)
	if err != nil {
		log.Printf("Got error calling Unmarshal: %s", err)
		return *dnaData, err
	}
	err = ValidateDna(dnaData.Dna)
	if err != nil {
		log.Printf("Got error calling ValidateDna: %s", err)
		return *dnaData, err
	}
	return *dnaData, nil
}

func UpdateData(dnaData DnaData, svc dynamodbiface.DynamoDBAPI) error {
	err := SaveDna(dnaData, svc)
	if err != nil {
		return err
	}
	err = UpdateStats(dnaData.Type, svc)
	return err
}

func UpdateStats(dnaType string, svc dynamodbiface.DynamoDBAPI) error {
	input := CreateUpdateItemInput(dnaType)
	_, err := svc.UpdateItem(input)
	if err != nil {
		log.Printf("Got error calling UpdateItem: %s", err)
		return err
	}
	return nil
}

func CreateUpdateItemInput(dnaType string) *dynamodb.UpdateItemInput {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(STATS_TABLE),
		Key: map[string]*dynamodb.AttributeValue{
			"dna_type": {
				S: aws.String(dnaType),
			},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":inc": {
				N: aws.String("1"),
			},
		},
		UpdateExpression: aws.String("ADD type_count :inc"),
	}
	return input
}

// Tengo que hacer un handling del SaveDna si no lo hago async
func SaveDna(dnaData DnaData, svc dynamodbiface.DynamoDBAPI) error {
	av, err := dynamodbattribute.MarshalMap(dnaData)
	if err != nil {
		log.Printf("Got error calling MarshalMap: %s", err)
		return err
	}
	tableName := DNAS_TABLE
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err = svc.PutItem(input)
	if err != nil {
		log.Printf("Got error calling PutItem: %s", err)
		return err
	}
	return nil
}

func GetDynamoDBClient() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)
	return svc
}

func Respond(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}

func ValidateDna(dna []string) error {
	re := regexp.MustCompile(`^[ACGT]+$`)
	for _, s := range dna {
		if !re.MatchString(s) {
			return errors.New("the dna provided does not match a possible dna")
		}
	}
	return nil
}

func IsMutant(dna []string) bool {
	length := len(dna)
	sequences := 0
	for i := 0; i < length; i++ {
		for j := 0; j < length; j++ {
			isSequence := isSequence(i, j, dna)
			if isSequence {
				sequences++
				if sequences >= NECESSARY_SECUENCES {
					return true
				}
			}
		}
	}
	return false
}

func isSequence(i int, j int, dna []string) bool {
	return IsHorizontalSequence(i, j, dna) ||
		IsVerticalSequence(i, j, dna) ||
		IsDiagonalSequence(i, j, dna)
}

func IsHorizontalSequence(indexI int, indexJ int, dna []string) bool {
	if indexJ > len(dna)-NECESSARY_SECUENCE {
		return false
	}
	c := strings.Split(dna[indexI], "")[indexJ]
	for i := 1; i < NECESSARY_SECUENCE; i++ {
		if c != strings.Split(dna[indexI], "")[indexJ+i] {
			return false
		}
	}
	return true
}

func IsVerticalSequence(indexI int, indexJ int, dna []string) bool {
	if indexI > len(dna)-NECESSARY_SECUENCE {
		return false
	}
	c := strings.Split(dna[indexI], "")[indexJ]
	for i := 1; i < NECESSARY_SECUENCE; i++ {
		if c != strings.Split(dna[indexI+i], "")[indexJ] {
			return false
		}
	}
	return true
}

func IsDiagonalSequence(indexI int, indexJ int, dna []string) bool {
	checkHorizontal := indexJ <= len(dna)-NECESSARY_SECUENCE
	checkVertical := indexI <= len(dna)-NECESSARY_SECUENCE
	if !checkHorizontal || !checkVertical {
		return false
	}
	c := strings.Split(dna[indexI], "")[indexJ]
	for i := 1; i < NECESSARY_SECUENCE; i++ {
		if c != strings.Split(dna[indexI+i], "")[indexJ+i] {
			return false
		}
	}
	return true
}
