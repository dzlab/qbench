package bench

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

type Parser struct {
	generators map[string]RecordGenerator
	channels   map[string]<-chan []byte
}

func NewParser() *Parser {
	return &Parser{
		generators: make(map[string]RecordGenerator),
		channels:   make(map[string]<-chan []byte),
	}
}

type Definition map[interface{}]interface{}

/*
 * A structure for attributes
 */
type Attribute struct {
	Name     string
	Channel  <-chan []byte
	Optional float64 // get the probability for including this attribute
}

type Object struct {
	random     *rand.Rand
	Attributes []Attribute
}

func NewObject(attributes []Attribute) *Object {
	source := rand.NewSource(time.Now().UnixNano())
	return &Object{random: rand.New(source), Attributes: attributes}
}

/*
 * Generate a key-value string of this object's attributes using the provided separator
 */
func (this *Object) GetKV(sep string) string {
	result := ""
	for _, a := range this.Attributes {
		// check if we have to ignore this attribute or not
		if a.Optional > 0 && this.random.Float64() > a.Optional {
			continue
		}
		value := string(<-a.Channel)
		result += a.Name + ":" + value + sep
	}
	result = result[:len(result)-len(sep)]
	return result
}

/*
 * Parse generator definitions
 */
func (this *Parser) parseGenerators(objects []interface{}) {
	var generator RecordGenerator
	var err error
	for _, elm := range objects {
		o := elm.(Definition)
		name := o["name"].(string)
		switch o["type"] {
		case "date":
			generator, err = NewDateGenerator(o["min"].(string), o["max"].(string), o["format"].(string))
		case "float":
			generator, err = NewFloatGenerator(o["min"].(float64), o["max"].(float64))
		case "pick":
			if o["file"] != nil {
				generator, err = NewPickFromFileGenerator(o["file"].(string))
			} else if o["values"] != nil {
				var values []string
				for _, v := range o["values"].([]interface{}) {
					values = append(values, v.(string))
				}
				generator, err = NewPickFromValuesGenerator(values)
			}
		case "increment":
			generator, err = NewIncrementGenerator(o["initial"].(int))
		}
		if err == nil {
			this.generators[name] = generator
			this.channels[name] = generator.Generate()
		}
	}
}

/*
 * Parse schema definition
 */
func (this *Parser) parseSchema(objects []interface{}) *Object {
	var attributes []Attribute
	for _, elm := range objects {
		a := Attribute{}
		d := elm.(Definition)
		// parse `name` attribute
		a.Name = d["name"].(string)
		// parse `value` attribute should start with `$`
		value := d["value"].(string)[1:]
		// ignore this attribute if no channel is available
		if this.channels[value] == nil {
			log.Printf("Ignoring attribute '%s' with no corresponding channel\n", a.Name)
			continue
		}
		a.Channel = this.channels[value]
		// parse `optional` attribute
		if d["optional"] != nil {
			a.Optional = d["optional"].(float64)
		} else {
			a.Optional = -1
		}
		attributes = append(attributes, a)
	}
	return NewObject(attributes)
}

/*
 * Parse a definition file
 */
func (this *Parser) Parse(filename string) (*Object, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config []Definition //map[string]interface{}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, err
	}
	var object *Object
	// extract information
	for _, v1 := range config {
		for k2, v2 := range v1 {
			if k2 == "generators" {
				// parse data generators
				this.parseGenerators(v2.([]interface{}))
			} else if k2 == "schema" {
				// parse the schema of the output object
				object = this.parseSchema(v2.([]interface{}))
			} else {
				log.Println("Uknown key: " + k2.(string))
			}
		}
	}
	return object, nil
}
