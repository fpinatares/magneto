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

type dependencies struct {
	db dynamodbiface.DynamoDBAPI
}

func GetDynamoDBClient() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)
	return svc
}

func main() {
	svc := GetDynamoDBClient()
	d := dependencies{
		db: svc,
	}
	lambda.Start(d.GetStats)
}

func (d *dependencies) GetStats(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	stats, err := d.GetStatsFromDB()
	if err != nil {
		respondError(http.StatusInternalServerError)
	}

	var stat Stat
	err = stat.SetValues(stats)
	if err != nil {
		respondError(http.StatusInternalServerError)
	}
	return respondOk(stat)
}

func (d *dependencies) GetStatsFromDB() ([]map[string]*dynamodb.AttributeValue, error) {
	params := &dynamodb.ScanInput{
		TableName: aws.String("stats"),
	}
	result, err := d.db.Scan(params)
	if err != nil {
		log.Printf("Query API call failed: %s", err)
		return nil, err
	}
	return result.Items, nil
}

func (s *Stat) SetCount(dnaType string, count int) {
	if dnaType == HUMAN {
		s.Human = count
	} else {
		s.Mutant = count
	}
}

func (s *Stat) SetRatio(countA float64, countB float64) {
	s.Ratio = countA / countB
}

func (s *Stat) SetValues(items []map[string]*dynamodb.AttributeValue) error {
	for _, i := range items {
		item, err := ParseItem(i)
		if err != nil {
			return err
		}
		s.SetCount(item.DnaType, item.Count)
	}
	s.SetRatio(float64(s.Mutant), float64(s.Human))
	return nil
}

func ParseItem(item map[string]*dynamodb.AttributeValue) (StatDB, error) {
	stat := StatDB{}
	err := dynamodbattribute.UnmarshalMap(item, &stat)
	if err != nil {
		log.Printf("Got error unmarshalling: %s", err)
	}
	return stat, err
}

func respondError(status int) (events.APIGatewayProxyResponse, error) {
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
