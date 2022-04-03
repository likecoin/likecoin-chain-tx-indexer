package db

import "testing"

func TestCreateQuery(t *testing.T) {
	tables := []struct {
		input  ISCNRecordQuery
		output string
	}{
		{
			input: ISCNRecordQuery{
				ContentMetadata: &ContentMetadata{
					Keywords: "Cyberspace,EFF",
				},
			},
			output: `{"contentMetadata":{"keywords":"Cyberspace,EFF"}}`,
		},
		{
			input: ISCNRecordQuery{
				ContentMetadata: &ContentMetadata{
					Type: "Article",
				},
			},
			output: `{"contentMetadata":{"@type":"Article"}}`,
		},
		{
			input: ISCNRecordQuery{
				ContentFingerprints: []string{
					"fingerprints",
				},
			},
			output: `{"contentFingerprints":["fingerprints"]}`,
		},
		{
			input: ISCNRecordQuery{
				Stakeholders: []Stakeholder{
					{
						Entity: &Entity{
							Id: "John Perry Barlow",
						},
					},
				},
			},
			output: `{"stakeholders":[{"entity":{"id":"John Perry Barlow"}}]}`,
		},
		{
			input: ISCNRecordQuery{
				ContentMetadata: &ContentMetadata{
					Type: "Article",
				},
				Stakeholders: []Stakeholder{
					{
						Entity: &Entity{
							Id: "John Perry Barlow",
						},
					},
				},
			},
			output: `{"contentMetadata":{"@type":"Article"},"stakeholders":[{"entity":{"id":"John Perry Barlow"}}]}`,
		},
	}
	for _, v := range tables {
		body, err := v.input.Marshal()
		if err != nil {
			t.Error(err)
		}
		if string(body) != v.output {
			t.Errorf("Expect %s\ngot %s", v.output, string(body))
		}
		t.Log(string(body))

	}
}
