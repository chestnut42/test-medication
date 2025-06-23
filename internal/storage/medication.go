package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/chestnut42/test-medication/internal/model"
)

// wrappedMedication is an internal equivalent of Medication.
// The purpose is to abstract away partition and sort keys as they
// tend to have implementation dependent format.
type wrappedMedication struct {
	PartitionKey string `dynamodbav:"PK"`
	SortKey      string `dynamodbav:"SK"`
	model.Medication
}

func wrap(m model.Medication) wrappedMedication {
	// TODO: figure out sort key. Very likely these medications are going to be queried by client/person or
	// partner/company + some pagination. Ideally we should figure out as much as possible at the start for form
	// a proper SortKey.
	// I put id as SortKey. Usually it helps to have SortKey even if you don't have one. Up to my knowledge:
	// if later you will need to add SortKey you can just run over your data on the fly. If you don't have it
	// on the schema - you'll have to re-create a table which can be a much bigger issue.
	return wrappedMedication{
		PartitionKey: m.Id,
		SortKey:      m.Id,
		Medication:   m,
	}
}

func (s *Service) CreateMedication(ctx context.Context, medication model.Medication) error {
	wrapped := wrap(medication)

	item, err := attributevalue.MarshalMap(wrapped)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	cond := expression.Name("Id").AttributeNotExists()

	expr, err := expression.NewBuilder().
		WithCondition(cond).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build expression: %w", err)
	}

	if _, err = s.database.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:                 aws.String(s.cfg.MedicationTable),
		Item:                      item,
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}); err != nil {
		var cfe *types.ConditionalCheckFailedException
		if errors.As(err, &cfe) {
			return fmt.Errorf("medication %s already exists: %w", medication.Id, ErrAlreadyExists)
		}
		return fmt.Errorf("failed to put item: %w", err)
	}
	return nil
}
