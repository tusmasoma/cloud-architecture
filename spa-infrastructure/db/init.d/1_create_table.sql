USE goTodoAPIDB;

DROP TABLE IF EXISTS todos;

CREATE TABLE todos (
    id CHAR(36) PRIMARY KEY,      -- UUIDなどを使う場合、CHAR(36)を使用
    completed BOOLEAN NOT NULL,   -- 完了状態
    body TEXT NOT NULL            -- タスクの内容
);
