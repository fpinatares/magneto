package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const NECESSARY_SECUENCE = 4
const NECESSARY_SECUENCES = 2
const MUTANT = "Mutant"
const HUMAN = "Human"

type Stat struct {
	Mutant int     `json:"count_mutant_dna"`
	Human  int     `json:"count_human_dna"`
	Ratio  float64 `json:"ratio"`
}

type StatDB struct {
	DnaType string `json:"dna_type"`
	Count   int    `json:"type_count"`
}

func main() {
	lambda.Start(GetStats)
}

func GetStats(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse and validate body
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	params := &dynamodb.ScanInput{
		TableName: aws.String("stats"),
	}
	result, err := svc.Scan(params)
	if err != nil {
		log.Printf("Query API call failed: %s", err)
	}

	var stat Stat
	for _, i := range result.Items {
		item := StatDB{}

		err = dynamodbattribute.UnmarshalMap(i, &item)
		log.Print(item)
		if err != nil {
			log.Printf("Got error unmarshalling: %s", err)
			return respond(http.StatusInternalServerError)
		}

		if item.DnaType == HUMAN {
			stat.Human = item.Count
		} else {
			stat.Mutant = item.Count
		}
		stat.Ratio = float64(stat.Mutant) / float64(stat.Human)
	}
	log.Print("Finalizando")
	return respondOk(stat)
}

func respond(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}

func respondOk(stat Stat) (events.APIGatewayProxyResponse, error) {
	bytes, _ := json.Marshal(stat)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(bytes),
	}, nil
}
