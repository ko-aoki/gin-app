package service

import (
	"context"
	"fmt"

	"gin-app/entity"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

func writeUsingDML(db string) error {
	ctx := context.Background()

	fmt.Printf("1\n")

	client, err := spanner.NewClient(ctx, db)
	if err != nil {

		fmt.Printf("%s\n", err)
		return err
	}
	defer client.Close()
	fmt.Printf("2\n")

	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL: `INSERT Singers (SingerId, FirstName, LastName) VALUES
				(12, 'Melissa', 'Garcia'),
				(13, 'Russell', 'Morales'),
				(14, 'Jacqueline', 'Long'),
				(15, 'Dylan', 'Shaw')`,
		}
		fmt.Printf("3\n")
		rowCount, err := txn.Update(ctx, stmt)
		fmt.Printf("4\n")
		if err != nil {
			fmt.Printf("%s", err)
			return err
		}
		fmt.Printf("%d record(s) inserted.\n", rowCount)
		return err
	})
	fmt.Printf("%s", err)
	return err
}

func getSandboxList(db string) ([]entity.Sandbox, error) {
	ctx := context.Background()

	fmt.Printf("1\n")

	client, err := spanner.NewClient(ctx, db)
	if err != nil {

		fmt.Printf("%s\n", err)
		return []entity.Sandbox{}, err
	}
	defer client.Close()
	fmt.Printf("2\n")

	var list []entity.Sandbox
	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		fmt.Printf("2_2\n")
		stmt := spanner.Statement{
			SQL: `SELECT 
				IntCl,
				StrCL, 
				ByteCl,
				BoolCl, 
				DateCl, 
				TimeStampCl, 
				JsonCl
			FROM 
				Sandbox`,
		}
		fmt.Printf("3\n")
		iter := txn.Query(ctx, stmt)
		defer iter.Stop()
		for {
			row, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				fmt.Printf("error1 %s", err)
				return err
			}
			entity := entity.Sandbox{}
			if err := row.Columns(
				&entity.IntCl,
				&entity.StrCl,
				&entity.ByteCl,
				&entity.BoolCl,
				&entity.DateCl,
				&entity.TimeStampCl,
				&entity.JsonCl); err != nil {
				fmt.Printf("error2 %s", err)
				return err
			}
			fmt.Printf("%d %s %s %v %s %s %s\n",
				entity.IntCl, entity.StrCl, entity.ByteCl, entity.BoolCl, entity.DateCl, entity.TimeStampCl, entity.JsonCl)
			list = append(list, entity)
		}

		fmt.Printf("4\n")
		if err != nil {
			fmt.Printf("error %s", err)
			return err
		}
		fmt.Printf("5\n")
		return err
	})
	fmt.Printf("6\n")
	return list, nil
}
