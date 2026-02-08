package db

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
func setupTestDB(ctx context.Context) error {
	// 1. テーブルの作成 (DDL)
	// Sandboxエンティティの構造に基づいたスキーマ定義
	adminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return err
	}
	defer adminClient.Close()

	op, err := adminClient.UpdateDatabaseDdl(ctx, &adminpb.UpdateDatabaseDdlRequest{
		Database: DB_PATH,
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
        ),
		(
            2, 2, 'テスト2', CAST('abc2' as BYTES), FALSE, DATE '2025-03-16', TIMESTAMP '2025-03-16T01:02:03.123457Z', JSON '{"key": "key1"}'
        )`

	// トランザクション内でDMLを実行
	client, err := spanner.NewClient(ctx, DB_PATH)
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

func cleanupTestDB(ctx context.Context) error {

	client, err := spanner.NewClient(ctx, DB_PATH)
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

	// 1. クリーンアップ処理の登録
	// t.Cleanup()に渡された関数は、TestGetSandboxListが終了した際に実行されます
	t.Cleanup(func() {
		if err := cleanupTestDB(ctx); err != nil {
			t.Errorf("Test cleanup failed: %v", err)
		}
	})

	// 2. テストデータのセットアップを実行
	if err := setupTestDB(ctx); err != nil {
		t.Fatalf("Test setup failed: %v", err)
	}

	wantTime, _ := time.Parse(time.RFC3339, "2025-03-16T01:02:03.123456Z")
	wantDate := civil.Date{Year: 2025, Month: time.March, Day: 16}
	wantJSONValue := map[string]interface{}{
		"key": "key1",
	}

	want := entity.Sandbox{
		KeyCl:  1,
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

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantKeyCl int64 // 期待される KeyCl。複数レコードの場合は最初のもの
		wantCount int
	}{
		{
			name:      "KeyClでフィルタリング (1)",
			params:    map[string]interface{}{"KeyCl": int64(1)},
			wantKeyCl: 1,
			wantCount: 1,
		},
		{
			name:      "IntClでフィルタリング (1)",
			params:    map[string]interface{}{"IntCl": int64(1)},
			wantKeyCl: 1,
			wantCount: 1,
		},
		{
			name:      "StrClでフィルタリング ('テスト')",
			params:    map[string]interface{}{"StrCl": "テスト"},
			wantKeyCl: 1,
			wantCount: 1,
		},
		{
			name:      "BoolClでフィルタリング (TRUE)",
			params:    map[string]interface{}{"BoolCl": true},
			wantKeyCl: 1,
			wantCount: 1,
		},
		{
			name:      "複合条件でフィルタリング (IntCl=1 AND BoolCl=TRUE)",
			params:    map[string]interface{}{"IntCl": int64(2), "BoolCl": false},
			wantKeyCl: 2,
			wantCount: 1,
		},
		{
			name: "DateClでフィルタリング (2025-03-16)",
			// DateClのフィルタリングには civil.Date 型を使用
			params: map[string]interface{}{"DateCl": wantDate},
			// setupTestDBではKeyCl 1と2の両方がこの日付を持つ
			wantKeyCl: 1, // KeyCl 1 または 2 が取得されれば成功。順序は KeyCl の昇順を期待
			wantCount: 2, // 両方のレコードが一致するはず
		},
		{
			name:      "DateClでフィルタリング (一致しない日付)",
			params:    map[string]interface{}{"DateCl": civil.Date{Year: 2026, Month: time.January, Day: 1}},
			wantCount: 0,
		},
		{
			name:      "TimeClでフィルタリング",
			params:    map[string]interface{}{"TimeStampCl": wantTime},
			wantKeyCl: 1,
			wantCount: 1,
		},
		{
			name:      "一致しないパラメータでのフィルタリング (結果なし)",
			params:    map[string]interface{}{"IntCl": int64(999)},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entityList, err := GetSandboxList(tt.params)
			if err != nil {
				t.Fatalf("getSandboxList failed: %v", err)
			}

			if len(entityList) != tt.wantCount {
				t.Fatalf("期待するレコード数: %d, 実際のレコード数: %d", tt.wantCount, len(entityList))
			}

			if tt.wantCount > 0 {
				// 取得された最初のレコードの KeyCl をチェック
				if entityList[0].KeyCl != tt.wantKeyCl {
					t.Errorf("期待する KeyCl: %d, 実際の KeyCl: %d", tt.wantKeyCl, entityList[0].KeyCl)
				}

				if tt.wantKeyCl == 1 && !reflect.DeepEqual(entityList[0], want) {
					t.Errorf("レコード内容が一致しません。期待: %+v, 取得: %+v", want, entityList[0])
				}
			}
		})
	}
}

// if len(entityList) < 1 {
// 	t.Fatalf("not found")
// }

// for _, entity := range entityList {
// 	fmt.Printf("entity %d", entity.IntCl)
// 	if !reflect.DeepEqual(entity, want) {
// 		t.Fatalf("expected: %v, got: %v", want, entity)
// 	}
// }

//}
