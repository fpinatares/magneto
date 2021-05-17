package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/google/uuid"
)

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
}

type mockDynamoDBClientError struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClient) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return nil, nil
}

func (m *mockDynamoDBClientError) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return nil, errors.New("Update item error")
}

func (m *mockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, nil
}

func (m *mockDynamoDBClientError) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, errors.New("Put item error")
}

func TestUpdateStats(t *testing.T) {
	d := dependencies{
		db: &mockDynamoDBClient{},
	}
	err := d.UpdateStats(EnumDnaType.Mutant)
	if err != nil {
		t.Error("No error expected updating stats", err)
	}
}

func TestErrorOnUpdateStats(t *testing.T) {
	d := dependencies{
		db: &mockDynamoDBClientError{},
	}
	err := d.UpdateStats("")
	if err == nil {
		t.Error("Expected error while updating stats", err)
	}
}

func TestSaveDna(t *testing.T) {
	dnaData := new(DnaData)
	dnaData.Uuid = uuid.New().String()
	dnaData.Type = EnumDnaType.Mutant
	d := dependencies{
		db: &mockDynamoDBClient{},
	}
	err := d.SaveDna(*dnaData)
	if err != nil {
		t.Error("No error expected saving dna", err)
	}
}

func TestUpdateData(t *testing.T) {
	dnaData := new(DnaData)
	dnaData.Dna = []string{"AAXAAA", "XXXXXX", "XXXXXX", "XXXXXX", "XXXXXX", "XXXXXX"}
	dnaData.Type = EnumDnaType.Human
	dnaData.Uuid = uuid.New().String()
	d := dependencies{
		db: &mockDynamoDBClient{},
	}
	err := d.UpdateData(*dnaData)
	if err != nil {
		t.Error("No error expected updating data", err)
	}
}

func TestErrorParsingEmptyRequest(t *testing.T) {
	body := ""
	_, err := ParseRequest(body)
	if err == nil {
		t.Error("Expected error while parsing an empty request", err)
	}
}

func TestParseRequest(t *testing.T) {
	body := "{\"dna\":[\"ATGCGA\",\"CAGTGC\",\"TTATTT\",\"AGACGG\",\"GCGTCA\",\"TCACTG\"]}"
	_, err := ParseRequest(body)
	if err != nil {
		t.Error("No error expected while parsing request", err)
	}
}

func TestRespondOK(t *testing.T) {
	response, err := Respond(200)
	if response.Body != "OK" || err != nil {
		t.Error("OK response expected")
	}
}

func TestSave(t *testing.T) {
	var record events.SNSEventRecord
	record.SNS.Message = "{\"uuid\":\"098765yh-876h-98j7-0o9i-987654tyh65t\",\"dna\":[\"ATGCGA\",\"CAGTGC\",\"TTATTT\",\"AGACGG\",\"GCGTCA\",\"TCACTG\"],\"type\":\"Human\"}"
	records := []events.SNSEventRecord{record}

	event := events.SNSEvent{
		Records: records,
	}
	d := dependencies{
		db: &mockDynamoDBClient{},
	}
	err := d.Save(event)
	if err != nil {
		t.Error("No error expected", err)
	}
}

func TestErrorUpdatingItemOnSave(t *testing.T) {
	var record events.SNSEventRecord
	record.SNS.Message = "{\"uuid\":\"098765yh-876h-98j7-0o9i-987654tyh65t\",\"dna\":[\"ATGCGA\",\"CAGTGC\",\"TTATTT\",\"AGACGG\",\"GCGTCA\",\"TCACTG\"],\"type\":\"Human\"}"
	records := []events.SNSEventRecord{record}

	event := events.SNSEvent{
		Records: records,
	}
	d := dependencies{
		db: &mockDynamoDBClientError{},
	}
	err := d.Save(event)
	if err == nil {
		t.Error("Error expected while saving item")
	}
}

func TestErrorUpdatingEmptyItemOnSave(t *testing.T) {
	var record events.SNSEventRecord
	record.SNS.Message = ""
	records := []events.SNSEventRecord{record}

	event := events.SNSEvent{
		Records: records,
	}
	d := dependencies{
		db: &mockDynamoDBClient{},
	}
	err := d.Save(event)
	if err == nil {
		t.Error("Expected error parsing an empty request")
	}
}
