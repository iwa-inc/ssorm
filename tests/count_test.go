package tests

import (
	"cloud.google.com/go/spanner"
	"context"
	"github.com/iwa-inc/ssorm"
	"testing"
)

func TestCountReadWrite(t *testing.T) {
	url := "projects/spanner-emulator/instances/test/databases/test"
	ctx := context.Background()

	client, _ := spanner.NewClient(ctx, url)
	defer client.Close()

	var (
		singer *Singers
		count  int64
	)
	_, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		err := ssorm.Model(singer).Where("SingerId in ?", []int{12, 13, 14, 15}).Count(ctx, txn, &count)
		return err
	})

	if err != nil {
		t.Fatalf("Error happened when count singers, got %v", err)
	}
}

func TestCountReadOnly(t *testing.T) {
	url := "projects/spanner-emulator/instances/test/databases/test"
	ctx := context.Background()

	client, _ := spanner.NewClient(ctx, url)
	defer client.Close()

	rtx := client.ReadOnlyTransaction()
	defer rtx.Close()

	var (
		singer *Singers
		count  int64
	)
	err := ssorm.Model(singer).Where("SingerId in ?", []int{12, 13, 14, 15}).Count(ctx, rtx, &count)

	if err != nil {
		t.Fatalf("Error happened when count singers, got %v", err)
	}
}
