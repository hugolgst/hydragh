package configuration

import (
	"encoding/json"
	"io/ioutil"
)

// Configuration contains the configuration file data, parsed
var Configuration = GetConfiguration()

// Data simply contains the parameters of the configuration file
type Data struct {
	HydraURL        string `json:"hydraURL"`
	AppID           int    `json:"appID"`
	CertificatePath string `json:"certificatePath"`
	Project         string `json:"project"`
	Jobs            []Job  `json:"jobs"`
}

// Job contains the structure of a Job, including its inputs
type Job struct {
	Name            string  `json:"name"`
	ExpressionInput string  `json:"expressionInput"`
	ExpressionPath  string  `json:"expressionPath"`
	Inputs          []Input `json:"inputs"`
}

// Input contains the structure of a Jobset input
type Input struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// GetConfiguration retrieves the configuration from the file and parse it
func GetConfiguration() (configuration Data) {
	data, err := ioutil.ReadFile("../configuration.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &configuration)
	if err != nil {
		panic(err)
	}

	return
}
