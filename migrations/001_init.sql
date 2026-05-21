-- +goose Up
CREATE TABLE IF NOT EXISTS departments (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(200) NOT NULL,
    parent_id  BIGINT REFERENCES departments(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_departments_no_self_parent CHECK (id != parent_id)
);


CREATE UNIQUE INDEX IF NOT EXISTS uq_departments_name_root
    ON departments(name)
    WHERE parent_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_departments_name_parent
    ON departments(parent_id, name)
    WHERE parent_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS employees (
    id            BIGSERIAL PRIMARY KEY,
    department_id BIGINT NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    full_name     VARCHAR(200) NOT NULL,
    position      VARCHAR(200) NOT NULL,
    hired_at      DATE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS employees;
DROP INDEX IF EXISTS uq_departments_name_parent;
DROP INDEX IF EXISTS uq_departments_name_root;
DROP TABLE IF EXISTS departments;
