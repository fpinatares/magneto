package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
}

type mockDynamoDBClientError struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClient) Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	return &dynamodb.ScanOutput{}, nil
}

func (m *mockDynamoDBClientError) Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	return nil, errors.New("Scan error")
}

func TestSetRatio(t *testing.T) {
	s := Stat{
		Mutant: 1,
		Human:  2,
	}
	s.CalculateRatio()
	if s.Ratio != 0.5 {
		t.Error("Expected ratio to be 0.5, got:", s.Ratio)
	}
}

func TestSetRatio0Human(t *testing.T) {
	s := Stat{
		Mutant: 10,
		Human:  0,
	}
	s.CalculateRatio()
	if s.Ratio != 10 {
		t.Error("Expected ratio to be 10, got:", s.Ratio)
	}
}

func TestSetHumanCount(t *testing.T) {
	s := Stat{}
	s.SetCount(EnumDnaType.Human, 8)
	if s.Human != 8 {
		t.Error("Expected Human count to be 8, got:", s.Human)
	}
}

func TestSetMutantCount(t *testing.T) {
	s := Stat{}
	s.SetCount(EnumDnaType.Mutant, 10)
	if s.Mutant != 10 {
		t.Error("Expected Mutant count to be 10, got:", s.Mutant)
	}
}

func TestUnsupportedDnaTypeCount(t *testing.T) {
	s := Stat{}
	err := s.SetCount("Cyborg", 4)
	if err == nil {
		t.Error("Expected error when setting unsopported dna type count")
	}
}

func TestGetStatsFromDB(t *testing.T) {
	d := dependencies{
		db: &mockDynamoDBClient{},
	}
	_, err := d.GetStatsFromDB()
	if err != nil {
		t.Error("Not error Expected Getting Stats From DB", err)
	}
}

func TestRespondOK(t *testing.T) {
	stat := Stat{
		Human:  1,
		Mutant: 2,
		Ratio:  0.5,
	}
	response, err := RespondOk(stat)
	if response.StatusCode != 200 || err != nil {
		t.Error("200 http status expected")
	}
}

func TestRespondError(t *testing.T) {
	response, err := RespondError(500)
	if response.Body != "Internal Server Error" || err != nil {
		t.Error("Internal Server Error response expected")
	}
}

func TestParseItem(t *testing.T) {
	item := make(map[string]*dynamodb.AttributeValue)
	item["dna_type"] = &dynamodb.AttributeValue{S: aws.String("Human")}
	item["type_count"] = &dynamodb.AttributeValue{N: aws.String("1")}
	_, err := ParseItem(item)
	if err != nil {
		t.Error("Error expected while parsing empty item", err)
	}
}

func TestSetValues(t *testing.T) {
	item1 := make(map[string]*dynamodb.AttributeValue)
	item1["dna_type"] = &dynamodb.AttributeValue{S: aws.String("Human")}
	item1["type_count"] = &dynamodb.AttributeValue{N: aws.String("2")}
	item2 := make(map[string]*dynamodb.AttributeValue)
	item2["dna_type"] = &dynamodb.AttributeValue{S: aws.String("Mutant")}
	item2["type_count"] = &dynamodb.AttributeValue{N: aws.String("1")}
	s := Stat{}
	items := []map[string]*dynamodb.AttributeValue{item1, item2}
	err := s.SetValues(items)
	if err != nil {
		t.Error("No error expected while setting values", err)
	}
	if s.Ratio != 0.5 {
		t.Error("Expected ratio to be 0.5. Got:", s.Ratio)
	}
}

func TestGetStats(t *testing.T) {
	d := dependencies{
		db: &mockDynamoDBClient{},
	}
	response, _ := d.GetStats()
	if response.StatusCode != 200 {
		t.Error("200 - Ok http status code expected. Got:", response.StatusCode)
	}
}

func TestGetStatsInternalServerError(t *testing.T) {
	d := dependencies{
		db: &mockDynamoDBClientError{},
	}
	response, _ := d.GetStats()
	if response.StatusCode != 500 {
		t.Error("500 http status code expected. Got:", response.StatusCode)
	}
}
