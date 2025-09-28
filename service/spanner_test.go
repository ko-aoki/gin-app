package service

import (
	"context"
	"fmt"
	"gin-app/entity"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	adminpb "cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
)

// setupTestDB はテスト前にテーブル作成とデータ挿入を行うヘルパー関数
func setupTestDB(ctx context.Context, db string) error {
	// 1. テーブルの作成 (DDL)
	// Sandboxエンティティの構造に基づいたスキーマ定義
	adminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return err
	}
	defer adminClient.Close()

	op, err := adminClient.UpdateDatabaseDdl(ctx, &adminpb.UpdateDatabaseDdlRequest{
		Database: db,
		Statements: []string{
			`CREATE TABLE Sandbox (
				KeyCl       INT64 NOT NULL,
				IntCl       INT64,
				StrCl       STRING(MAX),
				ByteCl      BYTES(MAX),
				BoolCl      BOOL,
				DateCl      DATE,
				TimeStampCl TIMESTAMP,
				JsonCl      JSON,
			) PRIMARY KEY (KeyCl)`,
		},
	})
	if err != nil {
		// テーブルが既に存在する場合、エラーになる可能性がありますが、
		// 開発環境によっては無視しても良い場合があります。
		fmt.Printf("Warning: CREATE TABLE failed (may already exist): %s\n", err)
	}
	if op != nil {
		if err := op.Wait(ctx); err != nil {
			return err
		}
	}

	// 2. データの挿入 (DML)
	// TestGetSandboxListで期待する値と一致させる
	insertDML := `
        INSERT INTO Sandbox (
            KeyCl, IntCl, StrCl, ByteCl, BoolCl, DateCl, TimeStampCl, JsonCl
        ) VALUES (
            1, 1, 'テスト', CAST('abc' as BYTES), TRUE, DATE '2025-03-16', TIMESTAMP '2025-03-16T01:02:03.123456Z', JSON '{"key": "key1"}'
        )`

	// トランザクション内でDMLを実行
	client, err := spanner.NewClient(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to create spanner client: %w", err)
	}
	defer client.Close()
	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		_, err := txn.Update(ctx, spanner.Statement{SQL: insertDML})
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to execute INSERT DML: %w", err)
	}

	return nil
}

func cleanupTestDB(ctx context.Context, db string) error {

	client, err := spanner.NewClient(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to create spanner client for cleanup: %w", err)
	}
	defer client.Close()

	// データの削除 (DML)
	// DELETE文ですべての行を削除します
	deleteDML := `DELETE FROM Sandbox WHERE TRUE`

	// ReadWriteTransaction内でDMLを実行
	_, err = client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		_, err := txn.Update(ctx, spanner.Statement{SQL: deleteDML})
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to execute DELETE DML: %w", err)
	}

	return nil
}

// func TestWriteUsingDML(t *testing.T) {
// 	writeUsingDML("projects/emu/instances/emu/databases/emu")
// }

func TestGetSandboxList(t *testing.T) {
	ctx := context.Background()
	dbPath := "projects/emu/instances/emu/databases/emu"
	database := dbPath // const dbPath を利用

	// 1. クリーンアップ処理の登録
	// t.Cleanup()に渡された関数は、TestGetSandboxListが終了した際に実行されます
	t.Cleanup(func() {
		if err := cleanupTestDB(ctx, database); err != nil {
			t.Errorf("Test cleanup failed: %v", err)
		}
	})

	// 2. テストデータのセットアップを実行
	if err := setupTestDB(ctx, database); err != nil {
		t.Fatalf("Test setup failed: %v", err)
	}

	entityList, err := getSandboxList(dbPath)
	if err != nil {
		fmt.Printf("err %s", err)
	}

	wantTime, _ := time.Parse(time.RFC3339, "2025-03-16T01:02:03.123456Z")
	wantJSONValue := map[string]interface{}{
		"key": "key1",
	}

	want := entity.Sandbox{
		IntCl:  1,
		StrCl:  "テスト",
		ByteCl: []byte("abc"),
		BoolCl: true,
		DateCl: spanner.NullDate{
			Date:  civil.Date{Year: 2025, Month: time.March, Day: 16},
			Valid: true,
		},
		TimeStampCl: spanner.NullTime{
			Time:  wantTime,
			Valid: true,
		},
		JsonCl: spanner.NullJSON{
			Value: wantJSONValue,
			Valid: true,
		},
	}

	for _, entity := range entityList {
		fmt.Printf("entity %d", entity.IntCl)
		if !reflect.DeepEqual(entity, want) {
			t.Fatalf("expected: %v, got: %v", want, entity)
		}
	}

}
