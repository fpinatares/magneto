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
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
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

type DnaData struct {
	Uuid string   `json:"uuid"`
	Dna  []string `json:"dna"`
	Type string   `json:"type"`
}

type dependencies struct {
	notifier snsiface.SNSAPI
}

func main() {
	svc := GetSNSClient()
	d := dependencies{
		notifier: svc,
	}
	lambda.Start(d.DetectMutant)
}

func (d *dependencies) DetectMutant(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.Headers["content-type"] != "application/json" && req.Headers["Content-Type"] != "application/json" {
		return Respond(http.StatusNotAcceptable)
	}
	dnaData, err := ParseRequest(req.Body)
	if err != nil {
		return Respond(http.StatusBadRequest)
	}
	dnaData.Uuid = uuid.New().String()
	dnaData.Type = GetDnaType(dnaData.Dna)
	json, _ := json.Marshal(dnaData)

	message := string(json)
	messagePtr := &message

	topicArn := "arn:aws:sns:us-east-1:870963517916:save-dna"
	topicArnPtr := &topicArn

	_, err = d.notifier.Publish(&sns.PublishInput{
		Message:  messagePtr,
		TopicArn: topicArnPtr,
	})
	if err != nil {
		log.Print(err)
	}
	if dnaData.Type != EnumDnaType.Mutant {
		return Respond(http.StatusForbidden)
	}
	return Respond(http.StatusOK)
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

func GetSNSClient() *sns.SNS {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := sns.New(sess)
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
			isSequence := IsSequence(i, j, dna)
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

func IsSequence(i int, j int, dna []string) bool {
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
