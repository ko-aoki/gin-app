CREATE TABLE Singers (
                         SingerId   INT64 NOT NULL,
                         FirstName  STRING(1024),
                         LastName   STRING(1024),
                         SingerInfo BYTES(MAX),
                         FullName   STRING(2048) AS (
					ARRAY_TO_STRING([FirstName, LastName], " ")
				) STORED
) PRIMARY KEY (SingerId)

CREATE TABLE Sandbox (
                         keyCl INT64  NOT NULL,
                         IntCl   INT64,
                         StrCl  STRING(1024),
                         ByteCl BYTES(MAX),
                         BoolCl BOOL,
                         DateCl DATE,
                         TimeStampCl TIMESTAMP,
                         JsonCl JSON
) PRIMARY KEY (keyCl)

INSERT INTO Sandbox
(keyCl, IntCl, StrCl, ByteCl, BoolCl, DateCl, TimeStampCl, JsonCl)
VALUES
    (
        1,
        1,
        'テスト',
        CAST('abc' as BYTES),
        true,
        '2025-03-16',
        '2025-03-16T01:02:03.123456Z',
        JSON '{"key":"key1"}'
    )
