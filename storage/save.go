package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

const NECESSARY_SECUENCE = 4
const NECESSARY_SECUENCES = 2
const STATS_TABLE = "stats"
const DNAS_TABLE = "dnas"

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

type DnaData struct {
	Uuid string   `json:"uuid"`
	Dna  []string `json:"dna"`
	Type string   `json:"type"`
}

type dependencies struct {
	db dynamodbiface.DynamoDBAPI
}

func main() {
	svc := GetDynamoDBClient()
	d := dependencies{
		db: svc,
	}

	lambda.Start(d.Save)
}

func (d *dependencies) Save(event events.SNSEvent) error {
	dnaData, err := ParseRequest(event.Records[0].SNS.Message)
	if err != nil {
		return err
	}
	err = d.UpdateData(dnaData)
	if err != nil {
		return err
	}
	return nil
}

func ParseRequest(body string) (DnaData, error) {
	dnaData := new(DnaData)
	err := json.Unmarshal([]byte(body), &dnaData)
	if err != nil {
		log.Printf("Got error calling Unmarshal: %s", err)
		return *dnaData, err
	}
	return *dnaData, nil
}

func (d *dependencies) UpdateData(dnaData DnaData) error {
	err := d.SaveDna(dnaData)
	if err != nil {
		return err
	}
	err = d.UpdateStats(dnaData.Type)
	return err
}

func (d *dependencies) UpdateStats(dnaType string) error {
	input := CreateUpdateItemInput(dnaType)
	_, err := d.db.UpdateItem(input)
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

func (d *dependencies) SaveDna(dnaData DnaData) error {
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
	_, err = d.db.PutItem(input)
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
