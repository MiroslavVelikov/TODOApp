BEGIN;

CREATE TABLE IF NOT EXISTS list (
    id UUID NOT NULL PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    name VARCHAR(100) UNIQUE NOT NULL,
    created_at DATE NOT NULL
);

CREATE TABLE IF NOT EXISTS users_lists (
    username VARCHAR(100) NOT NULL,
    list_id UUID NOT NULL REFERENCES list(id) ON DELETE CASCADE,
    is_owner BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT user_list_connection UNIQUE (list_id, username)
);

CREATE TYPE status_type
AS ENUM('Undefined', 'Not Assigned', 'Assigned', 'In Progress', 'In Review', 'Completed');

CREATE TYPE priority_type
AS ENUM('Undefined', 'Low', 'Medium', 'High');

CREATE TABLE IF NOT EXISTS todo (
    id UUID NOT NULL PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'),
    name VARCHAR(100) NOT NULL,
    list_id UUID NOT NULL REFERENCES list(id) ON DELETE CASCADE,
    description VARCHAR(1024),
    deadline DATE NOT NULL,
    assignee VARCHAR(100) DEFAULT '',
    created_at DATE NOT NULL,
    priority priority_type NOT NULL,
    status status_type NOT NULL DEFAULT 'Not Assigned',
    CONSTRAINT todo_list_constraint UNIQUE (name, list_id)
);

CREATE OR REPLACE FUNCTION modify_time_field()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.created_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$
    LANGUAGE PLPGSQL;

CREATE TRIGGER modify_time_field_when_insert_list_trigger
    BEFORE INSERT ON list
    FOR EACH ROW
EXECUTE FUNCTION modify_time_field();

CREATE TRIGGER modify_time_field_when_insert_todo_trigger
    BEFORE INSERT ON todo
    FOR EACH ROW
EXECUTE FUNCTION modify_time_field();

CREATE INDEX users_lists_username_index
ON users_lists(username);

CREATE INDEX users_lists_list_id_index
ON users_lists(list_id);

CREATE INDEX todo_list_id_index
ON todo(list_id);

COMMIT;