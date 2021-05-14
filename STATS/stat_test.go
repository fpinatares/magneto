package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
}

// Fijarme si no lo tengo que borrar
func (m *mockDynamoDBClient) Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	return &dynamodb.ScanOutput{}, nil
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
