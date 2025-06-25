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

func getPartition(m model.Identity) string {
	return m.Owner + "#" + m.Id // We can just concatenate here as it never leaves the implementation
}

// TODO: figure out sort key. Very likely these medications are going to be queried by client/person or
// partner/company + some pagination. Ideally we should figure out as much as possible at the start for form
// a proper SortKey.
// I put id as SortKey. Usually it helps to have SortKey even if you don't have one. Up to my knowledge:
// if later you will need to add SortKey you can just run over your data on the fly. If you don't have it
// on the schema - you'll have to re-create a table which can be a much bigger issue.
func getSortKey(m model.Identity) string {
	return m.Id
}

func (s *Service) CreateMedication(ctx context.Context, medication model.Medication) error {
	wrapped := wrappedMedication{
		PartitionKey: getPartition(medication.Identity),
		SortKey:      getSortKey(medication.Identity),
		Medication:   medication,
	}

	item, err := attributevalue.MarshalMap(wrapped)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	cond := expression.Name("PK").AttributeNotExists().
		And(expression.Name("SK").AttributeNotExists())

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

func (s *Service) GetMedication(ctx context.Context, identity model.Identity) (model.Medication, error) {
	resp, err := s.database.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.cfg.MedicationTable),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: getPartition(identity)},
			"SK": &types.AttributeValueMemberS{Value: getSortKey(identity)},
		},
	})
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return model.Medication{}, fmt.Errorf("medication not found: %v, %w", identity, ErrNotFound)
		}
		return model.Medication{}, fmt.Errorf("failed to get item: %w", err)
	}
	if resp.Item == nil {
		return model.Medication{}, fmt.Errorf("medication not found: %v, %w", identity, ErrNotFound)
	}

	// TODO: also return NotFound if the object is marked as DELETED

	var item wrappedMedication
	if err = attributevalue.UnmarshalMap(resp.Item, &item); err != nil {
		return model.Medication{}, fmt.Errorf("failed to unmarshal item: %w", err)
	}
	return item.Medication, nil
}

func (s *Service) UpdateMedication(ctx context.Context, oldVersion string, medication model.Medication) (model.Medication, error) {
	// TODO: implement me
	// The operation must succeed ONLY if the old version is equal to one in DB
	// Roughly we can just PutItem with a condition to overwrite the whole object conditionally.
	// Once the logic become more sophisticate we can move to UpdateItem certain fields
	//
	// Ideally we should save each updated snapshot in a logging table. It will cost money, but saves a ton of time
	// on issue resolution. If money isn't a concern here (95% it is not) we should do so.
	return model.Medication{}, errors.New("not implemented")
}

func (s *Service) DeleteMedication(ctx context.Context, identity model.Identity) error {
	// TODO: implement me
	// Ideally we should just mark the object as DELETED and leave it as is. Unless:
	// a) We must conform some GDPR and must delete the data
	// b) The caller wants to create new object with the same id -> we can save deleted object in log table and delete from the
	// main table.
	return errors.New("not implemented")
}
