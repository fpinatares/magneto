package main

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

type mockSNSClient struct {
	snsiface.SNSAPI
}

func (m *mockSNSClient) Publish(input *sns.PublishInput) (*sns.PublishOutput, error) {
	return nil, nil
}

func TestDetectMutantWithNotAcceptableContentType(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{},
	}
	req.Headers["content-type"] = "application/xml"
	d := dependencies{
		notifier: &mockSNSClient{},
	}
	response, _ := d.DetectMutant(req)
	if response.StatusCode != 406 {
		t.Error("406 - Not Acceptable http status code expected. Got:", response.StatusCode)
	}
}

func TestDetectMutantWithEmptyBody(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{},
		Body:    "",
	}
	req.Headers["content-type"] = "application/json"
	d := dependencies{
		notifier: &mockSNSClient{},
	}
	response, _ := d.DetectMutant(req)
	if response.StatusCode != 400 {
		t.Error("400 - Bad Request http status code expected. Got:", response.StatusCode)
	}
}

func TestDetectNoMutantRespondsForbidden(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{},
		Body:    "{\"dna\":[\"ATGCGA\",\"CAGTGC\",\"TTATTT\",\"AGACGG\",\"GCGTCA\",\"TCACTG\"]}",
	}
	req.Headers["content-type"] = "application/json"
	d := dependencies{
		notifier: &mockSNSClient{},
	}
	response, _ := d.DetectMutant(req)
	if response.StatusCode != 403 {
		t.Error("403 - Forbidden http status code expected. Got:", response.StatusCode)
	}
}

func TestDetectMutantRespondsOk(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{},
		Body:    "{\"dna\":[\"ATGCGA\",\"CAGTGC\",\"TTATGT\",\"AGAAGG\",\"CCCCTA\",\"TCACTG\"]}",
	}
	req.Headers["content-type"] = "application/json"
	d := dependencies{
		notifier: &mockSNSClient{},
	}
	response, _ := d.DetectMutant(req)
	if response.StatusCode != 200 {
		t.Error("200 - Ok http status code expected. Got:", response.StatusCode)
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
