package db

import "testing"

func TestCreateQuery(t *testing.T) {
	tables := []struct {
		input  iscnRecordQuery
		output string
	}{
		{
			input: iscnRecordQuery{
				ContentMetadata: &contentMetadata{
					Keywords: "Cyberspace,EFF",
				},
			},
			output: `{"contentMetadata":{"keywords":"Cyberspace,EFF"}}`,
		},
		{
			input: iscnRecordQuery{
				ContentMetadata: &contentMetadata{
					Type: "Article",
				},
			},
			output: `{"contentMetadata":{"@type":"Article"}}`,
		},
		{
			input: iscnRecordQuery{
				ContentFingerprints: []string{
					"fingerprints",
				},
			},
			output: `{"contentFingerprints":["fingerprints"]}`,
		},
		{
			input: iscnRecordQuery{
				Stakeholders: []stakeholder{
					{
						Entity: &entity{
							Id: "John Perry Barlow",
						},
					},
				},
			},
			output: `{"stakeholders":[{"entity":{"id":"John Perry Barlow"}}]}`,
		},
		{
			input: iscnRecordQuery{
				ContentMetadata: &contentMetadata{
					Type: "Article",
				},
				Stakeholders: []stakeholder{
					{
						Entity: &entity{
							Id: "John Perry Barlow",
						},
					},
				},
			},
			output: `{"contentMetadata":{"@type":"Article"},"stakeholders":[{"entity":{"id":"John Perry Barlow"}}]}`,
		},
	}
	for _, v := range tables {
		body, err := marshallQuery(v.input)
		if err != nil {
			t.Error(err)
		}
		if string(body) != v.output {
			t.Errorf("Expect %s\ngot %s", v.output, string(body))
		}
		t.Log(string(body))

	}
}
