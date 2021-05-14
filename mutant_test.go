package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/google/uuid"
)

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
}

func (m *mockDynamoDBClient) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return nil, nil
}

func (m *mockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, nil
}

func TestUpdateStats(t *testing.T) {
	svc := &mockDynamoDBClient{}
	err := UpdateStats(EnumDnaType.Mutant, svc)
	if err != nil {
		t.Error("No error expected updating stats", err)
	}
}

func TestSaveDna(t *testing.T) {
	svc := &mockDynamoDBClient{}
	dnaData := new(DnaData)
	dnaData.Uuid = uuid.New().String()
	dnaData.Type = EnumDnaType.Mutant

	err := SaveDna(*dnaData, svc)
	if err != nil {
		t.Error("No error expected saving dna", err)
	}
}

func TestUpdateData(t *testing.T) {
	dnaData := new(DnaData)
	dnaData.Dna = []string{"AAXAAA", "XXXXXX", "XXXXXX", "XXXXXX", "XXXXXX", "XXXXXX"}
	dnaData.Type = EnumDnaType.Human
	dnaData.Uuid = uuid.New().String()
	svc := &mockDynamoDBClient{}
	err := UpdateData(*dnaData, svc)
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

func TestInvalidDnaParsingRequest(t *testing.T) {
	body := "{\"dna\":[\"XTGAGA\",\"CAGTGC\",\"TTATTT\",\"AGACGG\",\"GCGTCA\",\"TCACTG\"]}"
	_, err := ParseRequest(body)
	if err == nil {
		t.Error("Expected error while parsing request with an invalid dna", err)
	}
}

func TestGetMutantDnaType(t *testing.T) {
	dna := []string{"ATGCGA", "CAGTGC", "TTATGT", "AGAAGG", "CCCCTA", "TCACTG"}
	dnaType := GetDnaType(dna)
	if dnaType != EnumDnaType.Mutant {
		t.Error("Expected mutant dnaType")
	}
}

func TestGetHumanDnaType(t *testing.T) {
	dna := []string{"ATGCGA", "CAGTGC", "TTATTT", "AGACGG", "GCGTCA", "TCACTG"}
	dnaType := GetDnaType(dna)
	if dnaType != EnumDnaType.Human {
		t.Error("Expected human dnaType")
	}
}

func TestNotHorizontalSecuence(t *testing.T) {
	dna := []string{"AAXAAA", "XXXXXX", "XXXXXX", "XXXXXX", "XXXXXX", "XXXXXX"}
	if IsHorizontalSequence(0, 0, dna) {
		t.Error("No Horizontal Sequence Expected")
	}
}

func Test00IsHorizontalSecuence(t *testing.T) {
	dna := []string{"AAAAAA", "XXXXXX", "XXXXXX", "XXXXXX", "XXXXXX", "XXXXXX"}
	if !IsHorizontalSequence(0, 0, dna) {
		t.Error("Expected Horizontal Sequence for positions (0,0) (0,1) (0,2) (0,3)")
	}
}

func TestNotVerticalSecuence(t *testing.T) {
	dna := []string{"AXXXXX", "AXXXXX", "AXXXXX", "XAXXXX", "XXXXXX", "XXXXX"}
	if IsVerticalSequence(0, 0, dna) {
		t.Error("No Vertical Sequence Expected")
	}
}

func Test00IsVerticalSecuence(t *testing.T) {
	dna := []string{"AXXXXX", "AXXXXX", "AXXXXX", "AXXXXX", "XXXXXX", "XXXXX"}
	if !IsVerticalSequence(0, 0, dna) {
		t.Error("Expected Horizontal Sequence for positions (0,0) (1,0) (2,0) (3,0)")
	}
}

func TestNotDiagonalSecuence(t *testing.T) {
	dna := []string{"XXXXXX", "XAXXXX", "XXAXXX", "XXXAXX", "XXXXXX", "XXXXX"}
	if IsDiagonalSequence(0, 0, dna) {
		t.Error("No Diagonal Sequence Expected")
	}
}

func Test00IsDiagonalSecuence(t *testing.T) {
	dna := []string{"AXXXXX", "XAXXXX", "XXAXXX", "XXXAXX", "XXXXXX", "XXXXX"}
	if !IsDiagonalSequence(0, 0, dna) {
		t.Error("Expected Diagonal Sequence for positions (0,0) (1,1) (2,2) (3,3)")
	}
}

func Test11IsDiagonalSecuence(t *testing.T) {
	dna := []string{"XXXXXX", "XAXXXX", "XXAXXX", "XXXAXX", "XXXXAX", "XXXXXX"}
	if !IsDiagonalSequence(1, 1, dna) {
		t.Error("Expected Diagonal Sequence for positions (1,1) (2,2) (3,3) (4,4)")
	}
}

func Test22IsDiagonalSecuence(t *testing.T) {
	dna := []string{"XXXXXX", "XXXXXX", "XXAXXX", "XXXAXX", "XXXXAX", "XXXXXA"}
	if !IsDiagonalSequence(2, 2, dna) {
		t.Error("Expected Diagonal Sequence for positions (2,2) (3,3) (4,4) (5,5)")
	}
}

func TestNoValidDna(t *testing.T) {
	dna := []string{"XXXXXX", "XXXXXX", "XXAXXX", "XXXAXX", "XXXXAX", "XXXXXA"}
	err := ValidateDna(dna)
	if err == nil {
		t.Error("Expected an invalid DNA")
	}
}

func TestValidDna(t *testing.T) {
	dna := []string{"ATGCGA", "CAGTGC", "TTATTT", "AGACGG", "GCGTCA", "TCACTG"}
	err := ValidateDna(dna)
	if err != nil {
		t.Error("Expected a valid DNA", err)
	}
}

func TestRespondOK(t *testing.T) {
	response, err := Respond(200)
	if response.Body != "OK" || err != nil {
		t.Error("OK response expected")
	}
}

func TestNotMutant(t *testing.T) {
	dna := []string{"ATGCGA", "CAGTGC", "TTATTT", "AGACGG", "GCGTCA", "TCACTG"}
	if IsMutant(dna) {
		t.Error("No mutant DNA expected")
	}
}

func TestIsMutant(t *testing.T) {
	dna := []string{"ATGCGA", "CAGTGC", "TTATGT", "AGAAGG", "CCCCTA", "TCACTG"}
	if !IsMutant(dna) {
		t.Error("Expected a mutant DNA")
	}
}
