package action

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/hokaccha/go-prettyjson"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/yaml"
)

func isDNSSubdomain(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("Cannot enforce DNS syntax validation on response of type %T", val)
	}

	validation.IsDNS1123Subdomain(str)
	if msgs := validation.IsDNS1123Subdomain(str); len(msgs) != 0 {
		return fmt.Errorf("%s", strings.Join(msgs, ", "))
	}

	return nil
}

func isYAML(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("Cannot enforce YAML syntax validation on response of type %T", val)
	}

	out := map[string]interface{}{}
	return yaml.Unmarshal([]byte(str), &out)
}

func toString(in bool) string {
	if in {
		return "true"
	}
	return "false"
}

func namespaceQuestion() *survey.Question {
	return &survey.Question{
		Name: "namespace",
		Prompt: &survey.Input{
			Message: "Please type Action namespace: ",
			Default: DefaultNamespace,
		},
		Validate: survey.ComposeValidators(survey.Required),
	}
}

func actionNameQuestion(defaultName string) *survey.Question {
	return &survey.Question{
		Name: "name",
		Prompt: &survey.Input{
			Message: "Please type Action name: ",
			Default: defaultName,
		},
		Validate: survey.ComposeValidators(survey.Required, isDNSSubdomain),
	}
}

func printJSON(in interface{}, w io.Writer) error {
	out, err := prettyjson.Marshal(in)
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	return err
}

func printYAML(in interface{}, w io.Writer) error {
	out, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}

// generateDNSName returns a DNS-1123 subdomain compliant random name
func generateDNSName() string {
	rand.Seed(time.Now().UnixNano())
	return strings.Replace(namesgenerator.GetRandomName(0), "_", "-", 1)
}