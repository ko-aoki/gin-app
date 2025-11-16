package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"gin-app/entity"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

const DB_PATH = "projects/emu/instances/emu/databases/emu"

func writeUsingDML() error {
	ctx := context.Background()

	client, err := spanner.NewClient(ctx, DB_PATH)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL: `INSERT Singers (SingerId, FirstName, LastName) VALUES
				(12, 'Melissa', 'Garcia'),
				(13, 'Russell', 'Morales'),
				(14, 'Jacqueline', 'Long'),
				(15, 'Dylan', 'Shaw')`,
		}
		rowCount, err := txn.Update(ctx, stmt)
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

func getSandboxList(param map[string]interface{}) ([]entity.Sandbox, error) {
	ctx := context.Background()

	client, err := spanner.NewClient(ctx, DB_PATH)
	if err != nil {
		return []entity.Sandbox{}, err
	}
	defer client.Close()

	// 1. テンプレート実行用のマップを作成
	// 元のparamのコピーと、Cndフラグを格納する
	templateParams := make(map[string]interface{}, len(param)*2)

	// 2. paramの内容を templateParams にコピーし、Cndフラグを追加
	// KeyClはフィルタリングに使われないと仮定し、スキップ
	for key, value := range param {
		// 元のフィルタリング値をそのままコピー (SpannerのParamsとして使用される)
		templateParams[key] = value

		// Cndフラグを追加。値が存在するため常にtrue
		cndKey := "Cnd" + key
		templateParams[cndKey] = true
	}

	tmpl := template.Must(template.New("getSandboxList").Parse(
		`SELECT 
		KeyCl,
		IntCl,
		StrCL, 
		ByteCl,
		BoolCl, 
		DateCl, 
		TimeStampCl, 
		JsonCl
	FROM 
		Sandbox
	WHERE
		1 = 1
	{{ if .CndIntCl }}
		AND
		IntCl = @IntCl
	{{ end }}
	{{ if .CndStrCl }}
		AND
		StrCL = @StrCL
	{{ end }}
	{{ if .CndByteCl }}
		AND
		ByteCL = @ByteCL
	{{ end }}
	{{ if .CndBoolCl }}
		AND
		BoolCL = @BoolCL
	{{ end }}
	{{ if .CndDateCl }}
		AND
		DateCL = @DateCL
	{{ end }}
	{{ if .CndTimeStampCl }}
		AND
		TimeStampCl = @TimeStampCl
	{{ end }}
	`))

	var sqlBuffer bytes.Buffer
	tmpl.Execute(&sqlBuffer, templateParams)
	sqlString := strings.TrimSpace(sqlBuffer.String())
	fmt.Printf("sql: %s\n", sqlString)

	var list []entity.Sandbox
	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{
			SQL:    sqlBuffer.String(),
			Params: param,
		}
		iter := txn.Query(ctx, stmt)
		defer iter.Stop()
		for {
			row, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			entity := entity.Sandbox{}
			if err := row.Columns(
				&entity.KeyCl,
				&entity.IntCl,
				&entity.StrCl,
				&entity.ByteCl,
				&entity.BoolCl,
				&entity.DateCl,
				&entity.TimeStampCl,
				&entity.JsonCl); err != nil {
				return err
			}
			fmt.Printf("%d %s %s %v %s %s %s\n",
				entity.IntCl, entity.StrCl, entity.ByteCl, entity.BoolCl, entity.DateCl, entity.TimeStampCl, entity.JsonCl)
			list = append(list, entity)
		}
		if err != nil {
			fmt.Printf("error %s", err)
			return err
		}
		return err
	})
	return list, nil
}
