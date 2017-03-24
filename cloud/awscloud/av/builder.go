package av

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Builder struct {
	tableName  string
	debug      bool
	err        error
	keys       map[string]*dynamodb.AttributeValue
	conditions []string
	adds       []string
	sets       []string
	removes    []string
	names      map[string]*string
	values     map[string]*dynamodb.AttributeValue
}

func NewBuilder(tableName string) *Builder {
	return &Builder{
		tableName: tableName,
	}
}

func (b *Builder) Key(field string, v *dynamodb.AttributeValue) {
	if b.keys == nil {
		b.keys = map[string]*dynamodb.AttributeValue{}
	}

	if v == nil {
		return
	}

	b.keys[field] = v
}

func (b *Builder) EQ(field string, v *dynamodb.AttributeValue) {
	b.apply(field, v, func(name, value string) {
		b.addCondition(name + " = " + value)
	})
}

func (b *Builder) NotEQ(field string, v *dynamodb.AttributeValue) {
	b.apply(field, v, func(name, value string) {
		b.addCondition("not (" + name + " = " + value + ")")
	})
}

func (b *Builder) Set(field string, v *dynamodb.AttributeValue) {
	b.apply(field, v, func(name, value string) {
		b.addSet(name + " = " + value)
	})
}

func (b *Builder) Remove(field string) {
	nameKey, _ := sanitize(field)

	b.addName(nameKey, field)
	b.addRemove(nameKey)
}

func (b *Builder) Increment(field string) {
	nameKey, _ := sanitize(field)
	b.addName(nameKey, field)
	b.addValue(":one", Int(1))
	b.addAdds(nameKey + " :one")
}

func (b *Builder) addCondition(v string) {
	if b.conditions == nil {
		b.conditions = []string{}
	}

	b.conditions = append(b.conditions, v)
}

func (b *Builder) addAdds(v string) {
	if b.adds == nil {
		b.adds = []string{}
	}

	b.adds = append(b.adds, v)
}

func (b *Builder) addSet(v string) {
	if b.sets == nil {
		b.sets = []string{}
	}

	b.sets = append(b.sets, v)
}

func (b *Builder) addRemove(v string) {
	if b.removes == nil {
		b.removes = []string{}
	}

	b.removes = append(b.removes, v)
}

func (b *Builder) addName(k, v string) {
	if b.names == nil {
		b.names = map[string]*string{}
	}
	b.names[k] = aws.String(v)
}

func (b *Builder) addValue(k string, v *dynamodb.AttributeValue) {
	if b.values == nil {
		b.values = map[string]*dynamodb.AttributeValue{}
	}
	if v != nil {
		b.values[k] = v
	}
}

func (b *Builder) apply(field string, v *dynamodb.AttributeValue, fn func(name, value string)) {
	if v == nil {
		return
	}

	nameKey, valueKey := sanitize(field)

	b.addName(nameKey, field)
	b.addValue(valueKey, v)

	fn(nameKey, valueKey)
}

func (b *Builder) Debug() {
	b.debug = true
}

func (b *Builder) buildConditionExpression() *string {
	if b.conditions == nil {
		return nil
	}

	return aws.String(strings.Join(b.conditions, " AND "))
}

func (b *Builder) buildUpdateExpression() *string {
	if b.adds == nil && b.sets == nil && b.removes == nil {
		return nil
	}

	expr := ""
	if b.adds != nil {
		expr += " ADD " + strings.Join(b.adds, ", ")
	}
	if b.sets != nil {
		expr += " SET " + strings.Join(b.sets, ", ")
	}
	if b.removes != nil {
		expr += " REMOVE " + strings.Join(b.removes, ", ")
	}

	return aws.String(expr)
}

func (b *Builder) BuildGetItemInput() (*dynamodb.GetItemInput, error) {
	if b.err != nil {
		return nil, b.err
	}

	return &dynamodb.GetItemInput{
		TableName:              aws.String(b.tableName),
		Key:                    b.keys,
		ConsistentRead:         aws.Bool(true),
		ReturnConsumedCapacity: aws.String("INDEXES"),
	}, nil
}

func (b *Builder) BuildDeleteItemInput() (*dynamodb.DeleteItemInput, error) {
	if b.err != nil {
		return nil, b.err
	}

	return &dynamodb.DeleteItemInput{
		TableName:                   aws.String(b.tableName),
		Key:                         b.keys,
		ConditionExpression:         b.buildConditionExpression(),
		ExpressionAttributeNames:    b.names,
		ExpressionAttributeValues:   b.values,
		ReturnConsumedCapacity:      aws.String("INDEXES"),
		ReturnItemCollectionMetrics: aws.String("SIZE"),
		ReturnValues:                aws.String("ALL_OLD"),
	}, nil
}

func (b *Builder) BuildPutItemInput(in interface{}) (*dynamodb.PutItemInput, error) {
	if b.err != nil {
		return nil, b.err
	}

	item, err := dynamodbattribute.MarshalMap(in)
	if err != nil {
		return nil, err
	}

	return &dynamodb.PutItemInput{
		Item:                        item,
		TableName:                   aws.String(b.tableName),
		ConditionExpression:         b.buildConditionExpression(),
		ExpressionAttributeNames:    b.names,
		ExpressionAttributeValues:   b.values,
		ReturnConsumedCapacity:      aws.String("INDEXES"),
		ReturnItemCollectionMetrics: aws.String("SIZE"),
		ReturnValues:                aws.String("NONE"),
	}, nil
}

func (b *Builder) BuildUpdateItemInput() (*dynamodb.UpdateItemInput, error) {
	if b.err != nil {
		return nil, b.err
	}

	in := &dynamodb.UpdateItemInput{
		TableName:                   aws.String(b.tableName),
		Key:                         b.keys,
		ConditionExpression:         b.buildConditionExpression(),
		UpdateExpression:            b.buildUpdateExpression(),
		ExpressionAttributeNames:    b.names,
		ExpressionAttributeValues:   b.values,
		ReturnConsumedCapacity:      aws.String("INDEXES"),
		ReturnItemCollectionMetrics: aws.String("SIZE"),
		ReturnValues:                aws.String("ALL_NEW"),
	}

	if b.debug {
		json.NewEncoder(os.Stdout).Encode(in)
	}

	return in, nil
}

var reSanitize = regexp.MustCompile(`[^a-zA-Z0-9]+(.)`)

func sanitize(v string) (name, value string) {
	matches := reSanitize.FindAllStringSubmatch(v, -1)
	if matches != nil {
		for _, match := range matches {
			v = strings.Replace(v, match[0], strings.ToUpper(match[1]), -1)
		}
	}

	name = "#" + v
	value = ":" + v
	return
}
