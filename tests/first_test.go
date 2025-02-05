package tests

import (
	"cloud.google.com/go/spanner"
	"context"
	"github.com/iwa-inc/ssorm"
	"testing"
)

func TestGetAllColumnReadWrite(t *testing.T) {
	url := "projects/spanner-emulator/instances/test/databases/test"
	ctx := context.Background()

	client, _ := spanner.NewClient(ctx, url)
	defer client.Close()

	singer := Singers{}
	//singer.TestTime = time.Now()
	_, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		err := ssorm.Model(&singer).Where("SingerId in ?", []int{12, 13, 14}).First(ctx, txn)
		return err
	})

	if err != nil {
		t.Fatalf("Error happened when get singer, got %v", err)
	}
}

func TestGetColumnReadWrite(t *testing.T) {
	url := "projects/spanner-emulator/instances/test/databases/test"
	ctx := context.Background()

	client, _ := spanner.NewClient(ctx, url)
	defer client.Close()

	singer := Singers{}
	_, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		err := ssorm.Model(&singer).Select([]string{"SingerId,FirstName"}).Where("SingerId in ?", []int{12, 13, 14}).First(ctx, txn)
		return err
	})

	if err != nil {
		t.Fatalf("Error happened when get singer, got %v", err)
	}
}

func TestGetAllColumnReadOnly(t *testing.T) {
	url := "projects/spanner-emulator/instances/test/databases/test"
	ctx := context.Background()

	client, _ := spanner.NewClient(ctx, url)
	defer client.Close()

	rtx := client.ReadOnlyTransaction()
	defer rtx.Close()

	singer := Singers{}
	err := ssorm.Model(&singer).Where("SingerId in ?", []int{12, 13, 14}).First(ctx, rtx)

	if err != nil {
		t.Fatalf("Error happened when get singer, got %v", err)
	}
}

func TestGetColumnReadOnly(t *testing.T) {
	url := "projects/spanner-emulator/instances/test/databases/test"
	ctx := context.Background()

	client, _ := spanner.NewClient(ctx, url)
	defer client.Close()

	rtx := client.ReadOnlyTransaction()
	defer rtx.Close()

	singer := Singers{}
	err := ssorm.Model(&singer).Select([]string{"SingerId,FirstName,LastName"}).
		Where("SingerId in ? and LastName = ? ",
			[]int{12, 13, 14},
			spanner.NullString{StringVal: "Morales", Valid: true},
		).First(ctx, rtx)

	if err != nil {
		t.Fatalf("Error happened when get singer, got %v", err)
	}
}
